"use strict";

/**
 * getAllEvents selects all events, and responds with event information
 * encoded as json
 * @param {Request} req HTTP request object 
 * @param {Response} res HTTP response object
 */

async function getAllEvents(req, res) {
  const db = req.db;
  const allEvents= [];
  try {
    const qry = "SELECT * FROM Events";
    const events = await db.query(qry);
    for (let i = 0; i < events.length; i++) {
        const event = events[i];
        const eventInfo = {
            "id": event.ID,
            "title": event.Title,
            "datetime": event.EventDateTime,
            "channel": event.ChannelID,
            "location": event.LocationOfEvent,
            "description": event.DescriptionOfEvent
          }
          allEvents.push(eventInfo);
    }
  } catch (err) {
    console.log(err.message);
    res.set("Content-Type", "text/plain");
    res.status(500).send("Server Error: Cannot select events in database.");
    db.end();
    return;
  }
  db.end();

  res.status(200).json(allEvents)
}

/**
 * createNewEvent creates new event using information
 * from event form. A new channel is also created that is
 * associated with this event.
 * @param {Request} req HTTP request object 
 * @param {Response} res HTTP response object
 */

async function createNewEvent(req, res) {
  const user = JSON.parse(req.get("X-User"));
  const event = req.body;
  const db = req.db;
  const createdAt = getDateTime();
  console.log("I am in");
  console.log(event.title);
  try {
    // Insert channel members into database if applicable
    const qryOne = "INSERT INTO Channels (ChannelName, ChannelDescription, PrivateChannel, TimeCreated, Creator, LastUpdated) VALUES (?,?,?,?,?,?);";
    // Insert creator of channel into member table
    const result = await db.query(qryOne, [event.title, event.description, false, createdAt, 1, createdAt])
    const channelID = result.insertId;
    const qryTwo = "INSERT INTO Events (Title, EventDateTime, ChannelID, LocationOfEvent, DescriptionOfEvent) VALUES (?,?,?,?,?);";
    await db.query(qryTwo, [event.title, event.datetime, channelID, event.location, event.description]);
  } catch (err) {
    console.log(err.message);
    res.set("Content-Type", "text/plain");
    res.status(500).send("Server Error: Cannot create new event or channel.");
    db.end();
    return;
  }
  db.end();

  res.status(200)
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
    getAllEvents,
    createNewEvent
}