"use strict";

const express = require("express");
const multer = require("multer");
const channels = require("./handlers/channels");
const messages = require("./handlers/messages");
const events = require("./handlers/events");
const auth = require("./middleware/auth");
const param = require("./middleware/param");
const db = require("./middleware/db");
const amqp = require('amqplib/callback_api');

const app = express();
const addr = process.env.ADDR || ":80";
const [host, port] = addr.split(":");

// ----- Middleware -----
// Checks if the user making a request to this microservice is authenticted
// (i.e. check the X-User header is set)
app.use(auth.isAuthenticatedUser);

// JSON parsing for application/x-www-form-urlencoded
app.use(express.urlencoded({ extended: true }))
// JSON parsing for application/json
app.use(express.json());
// JSON parsing for multipart/form-data (required with FormData)
app.use(multer().none());

// Establish a connection with the database and pass the connection 
// along in the request object
app.use(db.getDB);

// Establish a connection with RabbitMQ
amqp.connect('amqp://guest:guest@rabbitmqserver:5672/', function (error0, connection) {
  if (error0) {
    throw error0;
  }
  connection.createChannel(function (error1, channel) {
    if (error1) {
      throw error1;
    }
    var queue = 'events';

    channel.assertQueue(queue, {
      durable: true
    });
  });
});

// ----- API Routes -----
app.get("/v1/channels", channels.getAllChannels);
app.post("/v1/channels", channels.createNewChannel);

app.get("/v1/channels/:channelid", param.contains, auth.isMember, channels.getMessages);
app.post("/v1/channels/:channelid", param.contains, auth.isMember, channels.sendMessage);
app.patch("/v1/channels/:channelid", param.contains, auth.isCreator, channels.updateChannel);
app.delete("/v1/channels/:channelid", param.contains, auth.isCreator, channels.deleteChannel);

app.post("/v1/channels/:channelid/members", param.contains, auth.isCreator, channels.addUser);
app.delete("/v1/channels/:channelid/members", param.contains, auth.isCreator, channels.removeUser);

app.patch("/v1/messages/:messageid", param.contains, auth.isCreatorMsg, messages.updateMsg);
app.delete("/v1/messages/:messageid", param.contains, auth.isCreatorMsg, messages.deleteMsg);

app.listen(port, host, function () {
  console.log(`Server is listening at ${addr}...`);
});

