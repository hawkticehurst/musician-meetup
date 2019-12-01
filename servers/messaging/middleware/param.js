"use strict";

/**
 * contains middleware function that checks if a channel ID or message ID path parameter has been
 * included in the request. If a path parameter does not exist return a status 400 (Bad Request).
 * @param {Request} req HTTP request object
 * @param {Response} res HTTP response object
 * @param {NextFunction} next Function that should be called after this middleware
 */
function contains(req, res, next) {
  if (!req.params["channelid"] && !req.params["messageid"]) {
    res.set("Content-Type", "text/plain");
    res.status(400).send("Error: Missing path parameter.");
    return
  }
  next();
}

module.exports = {
  contains
}