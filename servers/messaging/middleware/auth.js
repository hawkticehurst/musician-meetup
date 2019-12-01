"use strict";

/**
 * isAuthenticatedUser middleware function that checks if the user making a request to this microservice 
 * is authenticted. The current user will be encoded as a JSON object in the X-User header. If that header 
 * is not in the request, assume that the user is unauthenticated and respond with status code 401 
 * (Unauthorized).
 * @param {Request} req HTTP request object
 * @param {Response} res HTTP response object
 * @param {NextFunction} next Function that should be called after this middleware
 */
function isAuthenticatedUser(req, res, next) {
  if (!req.get("X-User")) {
    res.set("Content-Type", "text/plain");
    res.status(401).send("Unauthorized");
    return
  }
  next();
}

/**
 * isMember middleware function that checks if the given channel is private and the current user is 
 * a member. If the current user is not a member of the channel respond with a 403 (Forbidden) 
 * status code.
 * @param {Request} req HTTP request object
 * @param {Response} res HTTP response object
 * @param {NextFunction} next Function that should be called after this middleware
 */
async function isMember(req, res, next) {
  const user = JSON.parse(req.get("X-User"));
  const channelID = req.params["channelid"];
  const result = await getSpecificChannelFromDB(channelID, req.db);
  if (result.error != null) {
    res.set("Content-Type", "text/plain");
    res.status(500).send("Server Error: Cannot get specified channel from database.");
    return;
  }
  const channel = result.channel;

  if (channel.private) {
    const members = channel.members;
    // This filters the channel members array and returns a new array containing channel members who
    // have the same ID as the current user (in this case there should be zero or one user in the array).
    // If the array is empty it means the current user is not a member of the channel and a 403 (Forbidden)
    // status code should be returned
    if (members.filter(memberID => memberID === user.id).length <= 0) {
      res.set("Content-Type", "text/plain");
      res.status(403).send("Forbidden: User is not a member of the specified channel.");
      return
    }
  }

  // Store user and channel JSON in request object so 
  // handler functions can later access this information
  req.user = user;
  req.channel = channel;
  next();
}

/**
 * isCreator middleware function that checks if the current user is the creator of the given
 * channel. If the current user is not the creator of the channel respond with a 403 (Forbidden) 
 * status code.
 * @param {Request} req HTTP request object
 * @param {Response} res HTTP response object
 * @param {NextFunction} next Function that should be called after this middleware
 */
async function isCreator(req, res, next) {
  const user = JSON.parse(req.get("X-User"));
  const channelID = req.params["channelid"];
  const result = await getSpecificChannelFromDB(channelID, req.db);
  if (result.error != null) {
    res.set("Content-Type", "text/plain");
    res.status(500).send("Server Error: Cannot get specified channel from database.");
    return;
  } 
  const channel = result.channel;

  if (user.id !== channel.creator) {
    res.set("Content-Type", "text/plain");
    res.status(403).send("Forbidden: User is not the creator of the specified channel.");
    return
  }

  // Store user and channel JSON in request object so 
  // handler functions can later access this information
  req.user = user;
  req.channel = channel;
  next();
}

/**
 * isCreatorMsg middleware function that checks if the current user is the creator of the given
 * message. If the current user is not the creator of the message respond with a 403 (Forbidden) 
 * status code.
 * @param {Request} req HTTP request object
 * @param {Response} res HTTP response object
 * @param {NextFunction} next Function that should be called after this middleware
 */
async function isCreatorMsg(req, res, next) {
  const user = JSON.parse(req.get("X-User"));
  const messageID = req.params["messageid"];
  const db = req.db;
  try {
    const qry = "SELECT * FROM Messages WHERE ID = ?";
    const result = await db.query(qry, [messageID]);
    const message = {
      "id": result[0].ID,
      "channelID": result[0].ChannelID,
      "body": result[0].Body,
      "createdAt": result[0].TimeCreated,
      "creator": result[0].Creator,
      "editedAt": result[0].LastUpdated
    };
    if (message.creator !== user.id) {
      res.set("Content-Type", "text/plain");
      res.status(403).send("Forbidden: User is not the creator of the specified message.");
      return
    }
    req.user = user;
    req.message = message;
  } catch (err) {
    console.log(err.message);
    res.set("Content-Type", "text/plain");
    res.status(500).send("Server Error: Cannot get specified message from database.");
    db.end();
    return;
  }
  // Note: We do not want to end the connection at this point because the handler function
  // called after this middleware will be using the same connection.
  next();
}

// ----- Helper Functions -----

/**
 * getSpecificChannelFromDB given a channel id, retrieve the associated channel from database and return
 * a JSON encoded version of the channel model 
 * @param {String} channelID A string representing a channel ID
 * @returns {JSON} A JSON object containing an encoded version of the specific channel model and 
 *                 a potential error
 */
async function getSpecificChannelFromDB(channelID, db) {
  let channel;
  try {
    const qry = "SELECT * FROM Channels WHERE ID = ?";
    const result = await db.query(qry, [channelID]);
    const membersQry = "SELECT * FROM ChannelsJoinMembers WHERE ChannelID = ?";
    const allMembers = await db.query(membersQry, [channelID]);
    var membersArray = [];
    for (let i = 0; i < allMembers.length; i++) {
      membersArray.push(allMembers[i].MemberID)
    }
    channel = {
      "id": result[0].ID,
      "name": result[0].ChannelName,
      "description": result[0].ChannelDescription,
      "private": result[0].PrivateChannel,
      "members": membersArray,
      "createdAt": result[0].TimeCreated,
      "creator": result[0].Creator,
      "editedAt": result[0].LastUpdated
    };
  } catch (err) {
    db.end();
    return { channel: null, error: err };
  }
  // Note: We do not want to end the DB connection at this point because the handler function
  // called after this middleware will be using the same connection.

  return { channel: channel, error: null };
}

module.exports = {
  isAuthenticatedUser,
  isMember,
  isCreator,
  isCreatorMsg
}