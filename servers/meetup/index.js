"use strict";

const express = require("express");
const multer = require("multer");
const events = require("./handlers/events");
const auth = require("./middleware/auth");
const db = require("./middleware/db");

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

// ----- API Routes -----
app.get("/v1/events", events.getAllEvents);
app.post("/v1/events", events.createNewEvent);
app.post("/v1/events/join", events.joinEvent);
app.get("/v1/events/join", events.getJoinedEvents);

app.listen(port, host, function () {
  console.log(`Server is listening at ${addr}...`);
});

