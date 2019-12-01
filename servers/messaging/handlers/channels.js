"use strict";

const rabbitmqhelpers = require("./rabbitmqhelpers");

/**
 * getAllChannels responds with the list of all channels (jut the channel models, 
 * not the messages in those channels) that the current user is allowed to see, 
 * encoded as a JSON array. Include a Content-Type header set to application/json 
 * so that your client knows what sort of data is in the response body.
 * @param {Request} req HTTP request object 
 * @param {Response} res HTTP response object
 */

async function getAllChannels(req, res) {
  const db = req.db;
  const user = JSON.parse(req.get("X-User"));
  const userChannels = [];
  try {
    // Find all private channels the user is a member of
    const qry = "SELECT * FROM ChannelsJoinMembers CM JOIN Channels C ON CM.ChannelID = C.ID WHERE CM.MemberID = ?";
    const privateChannels = await db.query(qry, [user.id]);
    for (let i = 0; i < privateChannels.length; i++) {
      const privateChannel = privateChannels[i];
      // Get all members profile
      const membersArray = await getMemberProfiles(privateChannel.ID, db);
      const privateChannelCreator = await getUserProfile(privateChannel.Creator, db);
      if (membersArray.error != null || privateChannelCreator.error != null) {
        res.set("Content-Type", "text/plain");
        res.status(500).send("Server Error: Cannot get profile information.");
        db.end();
        return;
      }
      const channelInfo = {
        "id": privateChannel.ID,
        "name": privateChannel.ChannelName,
        "description": privateChannel.ChannelDescription,
        "private": privateChannel.PrivateChannel,
        "members": membersArray.members,
        "createdAt": privateChannel.TimeCreated,
        "creator": privateChannelCreator.profile,
        "editedAt": privateChannel.LastUpdated
      }
      userChannels.push(channelInfo);
    }

    // Find and add all public channels to userChannels JSON
    const qryTwo = "SELECT * FROM Channels WHERE PrivateChannel = ?";
    const publicChannels = await db.query(qryTwo, [false]);
    for (let i = 0; i < publicChannels.length; i++) {
      const publicChannel = publicChannels[i];
      if (publicChannel.ChannelName != "General") {
        const publicChannelCreator = await getUserProfile(publicChannel.Creator, db);
        if (publicChannelCreator.error != null) {
          res.set("Content-Type", "text/plain");
          res.status(500).send("Server Error: Cannot get full creator profile from database.");
          db.end();
          return;
        }
        publicChannel.Creator = publicChannelCreator.profile;
        userChannels.push(publicChannel);
      } else {
        userChannels.push(publicChannel);
      }
    }
  } catch (err) {
    console.log(err.message);
    res.set("Content-Type", "text/plain");
    res.status(500).send("Server Error: Cannot get user channels from database.");
    db.end();
    return;
  }
  db.end();

  res.status(200).json(userChannels);
}

/**
 * createNewChannel creates a new channel using the channel model JSON in the request 
 * body. The name property is required, but description is optional. Respond with a 
 * 201 status code, a Content-Type set to application/json, and a copy of the new channel 
 * model (including its new ID) encoded as a JSON object.
 * @param {Request} req HTTP request object 
 * @param {Response} res HTTP response object
 */
async function createNewChannel(req, res) {
  const name = req.body.name;
  const createdAt = getDateTime();
  const creator = req.body.creator;
  let isPrivate = false;
  if (req.body.private) {
    isPrivate = req.body.private
  }
  let description = undefined;
  if (req.body.description) {
    description = req.body.description
  }
  let members = undefined;
  if (req.body.members) {
    members = req.body.members;
  }
  let editedAt = undefined;
  if (req.body.editedAt) {
    editedAt = new Date(req.body.editedAt);
  }

  const db = req.db;
  var channelID;
  try {
    // Insert new channel into database
    const qry = "INSERT INTO Channels (ChannelName, ChannelDescription, PrivateChannel, TimeCreated, Creator, LastUpdated) VALUES (?,?,?,?,?,?);";
    const result = await db.query(qry, [name, description, isPrivate, createdAt, creator.id, editedAt]);
    channelID = result.insertId;
    // Insert channel members into database if applicable
    const qryTwo = "INSERT INTO ChannelsJoinMembers (ChannelID, MemberID) VALUES (?,?);";
    // Insert creator of channel into member table
    await db.query(qryTwo, [channelID, creator.id])
    if (typeof members !== 'undefined') {
      for (let i = 0; i < members.length; i++) {
        await db.query(qryTwo, [channelID, members[i].id]);
      }
    }
  } catch (err) {
    console.log(err.message);
    res.set("Content-Type", "text/plain");
    res.status(500).send("Server Error: Cannot insert into database.");
    db.end();
    return;
  }

  const newChannelWithID = {
    "id": channelID,
    "name": name,
    "description": description,
    "private": isPrivate,
    "members": members,
    "createdAt": createdAt,
    "creator": creator,
    "editedAt": editedAt
  }

  const memberIDs = await rabbitmqhelpers.getMemberIDs(channelID, db);

  const newChannelObject = {
    "type": "channel-new",
    "channel": newChannelWithID,
    "userIDs": memberIDs.members
  }

  req.amqpChannel.sendToQueue("events", JSON.stringify(newChannelObject), { persistent: true });
  console.log(" [x] Sent %s", JSON.stringify(newChannelObject));

  res.status(201).json(newChannelWithID);
}

/**
 * getMessages responds with the most recent 100 messages posted to the specified
 * channel, encoded as a JSON array of message model objects. Include a Content-Type header set to 
 * application/json so that your client knows what sort of data is in the response body.
 * 
 * Add support for a before query string parameter that accepts a message ID. If provided, 
 * return the most recent 100 messages in the specified channel with message IDs less than the 
 * message ID in that query string parameter. This will allow a client to seamlessly request the 
 * previous page of results when the user scrolls back to the earliest message in the current page 
 * of results.
 * @param {Request} req HTTP request object 
 * @param {Response} res HTTP response object
 */
async function getMessages(req, res) {
  const channelID = req.params.channelid;
  const db = req.db;

  const before = req.params.before;
  let messages;

  if (before != undefined) {
    try {
      const qry = 'SELECT * FROM Messages WHERE ChannelID = ? AND ID < ? Order By TimeCreated Limit 100';
      messages = await db.query(qry, [channelID, before]);
    } catch (err) {
      console.log(err.message);
      res.set("Content-Type", "text/plain");
      res.status(500).send("Server Error: Cannot get message from database.");
      db.end();
      return;
    }
  } else {
    try {
      const qry = 'SELECT * FROM Messages WHERE ChannelID = ? Order By TimeCreated Limit 100';
      messages = await db.query(qry, [channelID]);
    } catch (err) {
      console.log(err.message);
      res.set("Content-Type", "text/plain");
      res.status(500).send("Server Error: Cannot get message from database.");
      db.end();
      return;
    }
  }

  const messagesList = []
  for (let i = 0; i < messages.length; i++) {
    const result = await getUserProfile(messages[i].Creator, db);
    if (result.error != null) {
      console.log(err.message);
      res.set("Content-Type", "text/plain");
      res.status(500).send("Server Error: Cannot get full creator profile from database.");
      db.end();
      return;
    }

    const msg = {
      "id": messages[i].ID,
      "channelID": messages[i].ChannelID,
      "body": messages[i].Body,
      "createdAt": messages[i].TimeCreated,
      "creator": result.profile,
      "editedAt": messages[i].LastUpdated
    }

    messagesList.push(msg);
  }
  db.end();

  res.set("Content-Type", "application/json");
  res.status(200).send(messagesList);
}

/**
 * sendMessage creates a new message in this channel using the JSON in the request
 * body. The only message property you should read from the request is body. Set the others based on 
 * context. Respond with a 201 status code, a Content-Type set to application/json, and a copy of the 
 * new message model (including its new ID) encoded as a JSON object.
 * @param {Request} req HTTP request object 
 * @param {Response} res HTTP response object
 */
async function sendMessage(req, res) {
  const user = req.user;
  const channelID = req.params.channelid;
  const db = req.db;

  let newMessage;
  try {
    const qry = 'INSERT INTO Messages (ChannelID, Body, TimeCreated, Creator, LastUpdated) VALUES (?, ?, NOW(), ?, NOW())';
    const qryResult = await db.query(qry, [channelID, req.body.body, user.id]);

    const qryTwo = 'SELECT * FROM Messages WHERE ID = ?';
    newMessage = await db.query(qryTwo, [qryResult.insertId]);
  } catch (err) {
    console.log(err.message);
  }

  const result = await getUserProfile(newMessage[0].Creator, db);
  if (result.error != null) {
    console.log(err.message);
    res.set("Content-Type", "text/plain");
    res.status(500).send("Server Error: Cannot get full creator profile from database.");
    db.end();
    return;
  }
  db.end();

  const newMsg = {
    "id": newMessage[0].ID,
    "channelID": newMessage[0].ChannelID,
    "body": newMessage[0].Body,
    "createdAt": newMessage[0].TimeCreated,
    "creator": result.profile,
    "editedAt": newMessage[0].LastUpdated
  }

  res.set("Content-Type", "application/json");

  const memberIDs = await rabbitmqhelpers.getMemberIDs(channelID, db);

  const sendMessageObject = {
    "type": "message-new",
    "message": newMsg,
    "userIDs": memberIDs.members
  }

  req.amqpChannel.sendToQueue("events", JSON.stringify(sendMessageObject), { persistent: true });
  console.log(" [x] Sent %s", JSON.stringify(sendMessageObject));

  res.status(201).json(newMsg);
}

/**
 * updateChannel updates only the name and/or description using the JSON in the request
 * body and respond with a copy of the newly-updated channel, encoded as a JSON object. Include 
 * a Content-Type header set to application/json so that your client knows what sort of data is 
 * in the response body.
 * @param {Request} req HTTP request object 
 * @param {Response} res HTTP response object
 */
async function updateChannel(req, res) {
  const channel = req.channel;
  let newName = undefined;
  if (req.body.name) {
    newName = req.body.name;
  }
  let newDescription = undefined;
  if (req.body.description) {
    newDescription = req.body.description;
  }

  // If neither a new name or description is given return a user error
  if (!newName && !newDescription) {
    res.set("Content-Type", "text/plain");
    res.status(400).send("Error: Please provide a new channel name or description in the request body.");
    return;
  }

  const db = req.db;
  const dateTime = getDateTime();
  try {
    if (typeof newName !== 'undefined') {
      const qry = "UPDATE Channels SET ChannelName = ?, LastUpdated = ? WHERE ID = ?;";
      await db.query(qry, [newName, dateTime, channel.id]);
      channel.name = newName;
      channel.editedAt = dateTime;
    }
    if (typeof newDescription !== 'undefined') {
      const qry = "UPDATE Channels SET ChannelDescription = ?, LastUpdated = ? WHERE ID = ?;";
      await db.query(qry, [newDescription, dateTime, channel.id]);
      channel.description = newDescription;
      channel.editedAt = dateTime;
    }
  } catch (err) {
    console.log(err.message);
    res.set("Content-Type", "text/plain");
    res.status(500).send("Server Error: Cannot update channel name and/or description in database.");
    db.end();
    return;
  }

  const result = await getUserProfile(channel.creator, db);
  if (result.error != null) {
    console.log(err.message);
    res.set("Content-Type", "text/plain");
    res.status(500).send("Server Error: Cannot get full creator profile from database.");
    db.end();
    return;
  }
  channel.creator = result.profile;
  if (channel.private) {
    const member = await getMemberProfiles(channel.id, db);
    if (member.error != null) {
      console.log(err.message);
      res.set("Content-Type", "text/plain");
      res.status(500).send("Server Error: Cannot get member profiles from database.");
      db.end();
      return;
    }
    channel.members = member.members;
  }

  const memberIDs = await rabbitmqhelpers.getMemberIDs(channel.id, db);

  const patchChannelObject = {
    "type": "channel-update",
    "channel": channel,
    "userIDs": memberIDs.members
  }

  req.amqpChannel.sendToQueue("events", JSON.stringify(patchChannelObject), { persistent: true });
  console.log(" [x] Sent %s", JSON.stringify(patchChannelObject));

  db.end();
  res.status(200).json(channel);
}

/**
 * deleteChannel deletes the channel and all messages related to it. Respond with a plain
 * text message indicating that the delete was successful.
 * @param {Request} req HTTP request object 
 * @param {Response} res HTTP response object
 */
async function deleteChannel(req, res) {
  const channel = req.channel;
  const db = req.db;

  try {
    // Delete channel from database
    const qry = "DELETE FROM Channels WHERE ID = ?;";
    await db.query(qry, [channel.id]);
    // Delete all messages related to the specified channel
    // If the table does not exist or there are no messages in the channel 
    // the resulting error will be ignored
    const qryTwo = "DELETE IGNORE FROM Messages WHERE ChannelID = ?;";
    await db.query(qryTwo, [channel.id]);
    // (If the channel is private) Delete all members related to the specified channel
    // If the table does not exist or there are no members in the channel 
    // the resulting error will be ignored
    if (channel.private) {
      const qryThree = "DELETE IGNORE FROM ChannelsJoinMembers WHERE ChannelID = ?;";
      await db.query(qryThree, [channel.id]);
    }
  } catch (err) {
    console.log(err.message);
    res.set("Content-Type", "text/plain");
    res.status(500).send("Server Error: Cannot delete channel or channel messages from database.");
    db.end();
    return;
  }

  const memberIDs = await rabbitmqhelpers.getMemberIDs(channel.id, db);

  const deleteChannelObject = {
    "type": "channel-delete",
    "channelID": channel.id,
    "userIDs": memberIDs.members
  }

  req.amqpChannel.sendToQueue("events", JSON.stringify(deleteChannelObject), { persistent: true });
  console.log(" [x] Sent %s", JSON.stringify(deleteChannelObject));

  db.end();
  res.status(200).type("text").send(`The ${channel.name} channel was successfully deleted.`);
}

/**
 * addUser adds the given user (supplied in the request body) as a member of this
 * channel, and responds with a 201 status code and a simple plain text message indicating that the 
 * user was added as a member. Only the id property of the user is required, but the client may post the 
 * entire user profile.
 * @param {Request} req HTTP request object 
 * @param {Response} res HTTP response object
 */
async function addUser(req, res) {
  const user = req.user;
  const channel = req.channel;
  const db = req.db;

  try {
    const qry = "INSERT INTO ChannelsJoinMembers (ChannelID, MemberID) VALUES (?, ?)";
    await db.query(qry, [user.id, channel.id]);
  } catch (err) {
    console.log(err.message);
    res.set("Content-Type", "text/plain");
    res.status(500).send("Server Error: Cannot add member to channel in database.");
    db.end();
    return;
  }
  db.end();

  res.status(201).type("text").send("User was added to the specified channel.");
}

/**
 * removeUser removes the user supplied in the request body from the list of channel
 * members, and respond with a 200 status code and a simple plain text message indicating that the user 
 * was removed from the list of members. Only the id property of the user is required, but the client 
 * may post the entire user profile.
 * @param {Request} req HTTP request object 
 * @param {Response} res HTTP response object
 */
async function removeUser(req, res) {
  const user = req.user;
  const channel = req.channel;
  const db = req.db;

  try {
    const qry = "DELETE FROM ChannelsJoinMembers WHERE ChannelID = ? AND MemberID = ?;";
    await db.query(qry, [user.id, channel.id]);
  } catch (err) {
    console.log(err.message);
    res.set("Content-Type", "text/plain");
    res.status(500).send("Server Error: Cannot remove member from database.");
    db.end();
    return;
  }
  db.end();

  res.status(200).type("text").send("User was removed from the specified channel.");
}

// ----- Helper Functions -----

// getUserProfile returns user information given the 
// user ID so the 'creator' field can be populated with 
// an entire profile
async function getUserProfile(userID, db) {
  try {
    const qry = "SELECT * FROM Users WHERE ID = ?";
    const user = await db.query(qry, [userID]);
    const userProfile = {
      "id": user[0].ID,
      "userName": user[0].UserName,
      "firstName": user[0].FirstName,
      "lastName": user[0].LastName,
      "photoURL": user[0].PhotoURL
    }
    return { profile: userProfile, error: null };
  } catch (err) {
    return { profile: null, error: err };
  }
}

//getMemberProfiles returns user profiles of all
//members of given channel
async function getMemberProfiles(channelID, db) {
  const membersArray = [];
  try {
    const qry = "SELECT ID, UserName, FirstName, LastName, PhotoURL FROM ChannelsJoinMembers C INNER JOIN Users U ON C.MemberID = U.ID AND C.ChannelID = ?";
    const user = await db.query(qry, [channelID]);
    for (let i = 0; i < user.length; i++) {
      const userProfile = {
        "id": user[i].ID,
        "userName": user[i].UserName,
        "firstName": user[i].FirstName,
        "lastName": user[i].LastName,
        "photoURL": user[i].PhotoURL
      }
      membersArray.push(userProfile);
    }
    return { members: membersArray, error: null };
  } catch (err) {
    return { members: null, error: err };
  }
}

function getDateTime() {
  const today = new Date();
  const date = today.getFullYear() + '-' + (today.getMonth() + 1) + '-' + today.getDate();
  const time = today.getHours() + ":" + today.getMinutes() + ":" + today.getSeconds();
  const dateTime = date + ' ' + time;
  return dateTime;
}
/**
 * Expose public handler functions.
 */
module.exports = {
  getAllChannels,
  createNewChannel,
  getMessages,
  sendMessage,
  updateChannel,
  deleteChannel,
  addUser,
  removeUser
}
