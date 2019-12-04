"use strict";

const mysql = require("promise-mysql");

/**
 * getDB middleware function that establishes a connection with the database and passes a reference to the
 * DB connection along in the Request object. If a connection fails to be created a 500 (Server Error) is
 * returned to the client.
 * @param {Request} req HTTP request object
 * @param {Response} res HTTP response object
 * @param {NextFunction} next Function that should be called after this middleware
 */
async function getDB(req, res, next) {
  try {
    // Create connection with database
    let db = await mysql.createConnection({
      host: process.env.HOST,
      port: process.env.PORT,
      user: process.env.USER,
      password: process.env.MYSQL_ROOT_PASSWORD,
      database: process.env.DATABASE,
      debug: true
    });
    // Store reference to DB connection in request object so 
    // handler functions can use the connection
    req.db = db;
    next();
  } catch (err) {
    console.error(err);
    res.set("Content-Type", "text/plain");
    res.status(500).send("Server Error: Cannot connect to database.");
    return
  }
}

module.exports = {
  getDB
}