"use strict";

const amqp = require('amqplib/callback_api');

/**
 * getRabbitMQConnection
 */
async function getRabbitMQConnection(req, res, next) {
  let amqpConn = null;
  let amqpChannel = null;

  // Create connection with RabbitMQ
  amqp.connect('amqp://guest:guest@rabbitmqserver:5672/', function (err0, connection) {
    if (err0) {
      console.error(err0);
      res.set("Content-Type", "text/plain");
      res.status(500).send("RabbitMQ Error: Cannot connect to rabbitMQ server.");
      return
    }

    connection.createChannel(function (err1, channel) {
      if (err1) {
        console.error(err1);
        res.set("Content-Type", "text/plain");
        res.status(500).send("RabbitMQ Error: Cannot create to rabbitMQ channel.");
        return
      }

      const queue = 'events';
      channel.assertQueue(queue, {
        durable: true
      });
      amqpChannel = channel;
    });

    amqpConn = connection;
  });

  console.log("AMQP Connection: " + amqpConn);
  console.log("AMQP Channel: " + amqpChannel);
  // Store reference to rabbitmq connection and channel
  req.amqpConn = amqpConn;
  req.amqpChannel = amqpChannel;
  next();
}

/**
 * Expose public helper functions.
 */
module.exports = {
  getRabbitMQConnection
}