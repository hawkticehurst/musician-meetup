"use strict";

const rabbitmqhelpers = require("./rabbitmqhelpers");

/**
 * updateMsg updates the message body property using the JSON in the request body, and respond with a 
 * copy of the newly-updated message, encoded as a JSON object. Include a Content-Type header set to 
 * application/json so that your client knows what sort of data is in the response body.
 * @param {Request} req HTTP request object 
 * @param {Response} res HTTP response object
 */
async function updateMsg(req, res) {
  const message = req.message;
  let newBody = req.body.body;
  const db = req.db;

  let editedAt;
  try {
    const qry = "UPDATE Messages SET Body = ?, LastUpdated = NOW() WHERE ID = ?;";
    await db.query(qry, [newBody, message.id]);
    const qryTwo = "SELECT * FROM Messages WHERE ID = ?;";
    const result = await db.query(qryTwo, [message.id]);
    editedAt = result[0].LastUpdated;
  } catch (err) {
    console.log(err.message);
    res.set("Content-Type", "text/plain");
    res.status(500).send("Server Error: Cannot update message in database.");
    db.end();
    return;
  }

  const result = await getUserProfile(message.creator, db);
  if (result.error != null) {
    res.set("Content-Type", "text/plain");
    res.status(500).send("Server Error: Cannot get full creator profile from database.");
    db.end();
    return;
  }
  db.end();

  const updatedMsg = {
    "id": message.id,
    "channelID": message.channelID,
    "body": newBody,
    "createdAt": message.createdAt,
    "creator": result.profile,
    "editedAt": editedAt
  }

  const memberIDs = await rabbitmqhelpers.getMemberIDs(message.channelID, db);

  const rabbitUpdateMessage = {
    "type": "message-update",
    "message": updatedMsg,
    "userIDs": memberIDs.members
  }

  const error = sendMessageToRabbitMQ(rabbitUpdateMessage);
  if (error != null) {
    console.log("Error sending event message to RabbitMQ: " + error);
    res.status(500).send("Error sending event message to RabbitMQ");
  }

  res.status(200).json(updatedMsg)
}

/**
 * deleteMsg if the current user isn't the creator of this message, respond with the status
 * code 403 (Forbidden). Otherwise, delete the message and respond with a the plain text message 
 * indicating that the delete was successful.
 * @param {Request} req HTTP request object 
 * @param {Response} res HTTP response object
 */
async function deleteMsg(req, res) {
  const message = req.message;
  const db = req.db;
  try {
    const qry = "DELETE FROM Messages WHERE ID = ?";
    await db.query(qry, [message.id]);
  } catch (err) {
    console.log(err.message);
    res.set("Content-Type", "text/plain");
    res.status(500).send("Server Error: Cannot delete message from database.");
    db.end();
    return;
  }
  db.end();

  const memberIDs = await rabbitmqhelpers.getMemberIDs(channel.id, db);

  const rabbitDeleteMessage = {
    "type": "message-delete",
    "channelID": message.id,
    "userIDs": memberIDs.members
  }

  const error = sendMessageToRabbitMQ(rabbitDeleteMessage);
  if (error != null) {
    console.log("Error sending event message to RabbitMQ: " + error);
    res.status(500).send("Error sending event message to RabbitMQ");
  }

  res.status(200).type("text").send("The message was deleted successfully.")
}

// ----- Helper Functions -----

// getUserProfile returns user information given the user ID so the 'creator' field 
// can be populated with an entire profile
async function getUserProfile(userID, db) {
  try {
    const qry = "SELECT ID, UserName, FirstName, LastName, PhotoURL FROM Users WHERE ID = ?;";
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

// sendMessageToRabbitMQ connects to RabbitMQ and sends a given message to the 'events' queue 
function sendMessageToRabbitMQ(message) {
  // Connect to RabbitMQ
  amqp.connect('amqp://guest:guest@rabbitmqserver:5672/', function (error0, connection) {
    if (error0) {
      return error0;
    }

    // Create a channel so we can send to the 'events' queue
    connection.createChannel(function (error1, channel) {
      if (error1) {
        return error1;
      }

      const queue = 'events';

      // Send the given message
      channel.sendToQueue(queue, Buffer.from(JSON.stringify(message)), { persistent: true });
      console.log("Sent to RabbitMQ: %s", JSON.stringify(message));
    });
  });

  return null;
}

/**
 * Expose public handler functions.
 */
module.exports = {
  updateMsg,
  deleteMsg
}