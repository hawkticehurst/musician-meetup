# INFO 441 Final Project Proposal

## Project Description

Our target audience is people who want to play music together. We want to build a platform that allows people to create music meetup events and facilitates musicians being able to enjoy playing music together. 

It can be hard for people playing an instrument to find people playing similar instruments or a group/band that may be looking for a player to join their group (i.e. a Jazz band needs a bass player). This platform will allow people to post public meetup events and message other players. This platform will also allow users to learn from other players nearby.

As developers, we are intrigued by the challenge that live-messaging with web sockets presents. We are also interested to see if we can create a seamless transition for users from chatting to creating meetups. 

## Technical Description

### Infrastructure

Users will interact exclusively with our web UI Docker container (hosted on AWS). The web interface container will interact with an API Gateway service (via a REST API) that will authenticate users and facilitate communication with all the other microservices in our backend. User, meetup, and chat information will be stored in a MySQL database container, and session information will be stored in a Redis database. We also use a RabbitMQ container as a queue that will notify each live WebSockets connection that a new chat message has been sent to any given chat.

### Service Architecture

![Diagram of the Service Architecture](./assets/architecture.png)

API Gateway is responsible for upgrading and storing CLIENT websocket connections.

### User Stories

| **Priority** | **User** | **Description** |
|--------------|----------|-----------------|
| P0 (High) | As a user | I want to be able to create a user account and log in |
| P0 (High) | As a user | I want to create public meetup events that other music players can discover |
| P0 (High) | As a user | I want to be able to view meetup events that other users have created |
| P0 (High) | As a user | I want to chat with other music players to plan meetups and discuss music together |
| P1 (Med) | As a user | I want to be able to join meetup events that have already been created |


**Story #1: I want to be able to create a user account and log in**

We will create a **Dockerized** **Go** web microservice that acts as an API gateway. This web service will expose a REST API (over port 443) that the Web UI can call. This gateway will facilitate user account creation and authentication.

The service will maintain a connection to our **MySQL** database (over port 3306) in order to save user information. The service will also maintain a connection to our **Redis** database (over port 6379) in order to create, track, and delete user sessions.

**Story #2: I want to create public meetup events that other music players can discover**

We will create a **Dockerized** **Node.js** web microservice for creating, storing, and deleting meetup events. This web service will expose a REST API (over port 80) that the API Gateway can call. 

The service will maintain a connection to our **MySQL** database (over port 3306) in order to store meetup information.

**Story #3: I want to be able to view meetup events that other users have created**

 This is an augmentation to the above **Dockerized** **Node.js** web microservice that will retrieve all meetup events from our **MySQL** database (over port 3306) and return them to the client to be displayed via the Web UI.

**Story #4: I want to chat with other music players to plan meetups and discuss music together**

We will create a **Dockerized** **Node.js** web microservice to facilitate messaging between users. This web service will expose a REST API (over port 80) that the API Gateway can call. This messaging service will maintain a connection to our **MySQL** database (over port 3306) to maintain messaging history.

When a new message is created, this microservice will notify a **RabbitMQ** queue of this event. The new message event will be consumed by the API gateway and write the contents of the new message to every live WebSocket connection that the gateway currently has stored.

**Story #5: I want to be able to join meetup events that are already created**

This will be an update to our **Dockerized** **Node.js** Meetup microservice. This update will implement a REST API PATCH update that will add the user to the given meetup.

The service will maintain a connection to our **MySQL** database (over port 3306) in order to store and update this information.

## API Endpoints

```/v1/users```
- ```POST```: post information to user database to create user account
    - ```201```: created new user
    - ```401```: Could not create user, or invalid session
    - ```415```: client did not post JSON
    - ```500```: Server error

```/v1/sessions```
- ```POST```: post information to user database to create user account
    - ```201```: create a new log in session
    - ```401```: Could not log user in
    - ```500```: Server error

```/v1/sessions/mine```
- ```DELETE```: post information to user database to create user account
    - ```200```: deleted the user session
    - ```401```: Could not delete given session
    - ```500```: Server error

```/v1/ws```
- Create websockets connection

```/v1/events```
- ```GET```: get all meetup events
    - ```201```: returns a list of all events
    - ```401```: Could not retreive events
    - ```500```: Server error
- ```POST```: post a newmeetup event
    - ```201```: create new meetup event
    - ```401```: could not create new meetup event
    - ```500```: Server error

```/v1/events/join```
- ```GET```: get information about events that the logged in user has joined
    - ```201```: returns list of channels the user is a part of
    - ```401```: could not get list of channels the user has joined
    - ```500```: Server error
- ```POST```: joins the logged in user to the event
    - ```201```: joins the user to the given channel
    - ```401```: could not join the user to the channel
    - ```500```: Server error

## Database Schemas

We will use MySQL as our persistent database.

Users: User represents a person who can log-in, message, and be a part of meetups on our site
```
CREATE TABLE IF NOT EXISTS Users (
    ID INT NOT NULL AUTO_INCREMENT,
    Email VARCHAR(255) NOT NULL UNIQUE,
    PassHash VARCHAR(72) NOT NULL,
    UserName VARCHAR(255) NOT NULL UNIQUE,
    FirstName VARCHAR(128),
    LastName VARCHAR(128),
    PhotoURL VARCHAR(2083) NOT NULL,
    PRIMARY KEY (ID)
);
```

Events: Represents an event that multiple users can join
```
CREATE TABLE IF NOT EXISTS Events (
    ID INT NOT NULL AUTO_INCREMENT,
    Title VARCHAR(255) NOT NULL,
    EventDateTime VARCHAR(255) NOT NULL,
    ChannelID INT NOT NULL,
    LocationOfEvent VARCHAR(255) NOT NULL,
    DescriptionOfEvent VARCHAR(255) NOT NULL,
    PRIMARY KEY (ID),
    FOREIGN KEY (ChannelID) REFERENCES Channels(ID)
);
```

Channels: Represents a chat channel
```
CREATE TABLE IF NOT EXISTS Channels (
    ID INT NOT NULL AUTO_INCREMENT,
    ChannelName VARCHAR(255) NOT NULL,
    ChannelDescription VARCHAR(255),
    PrivateChannel BOOLEAN NOT NULL,
    TimeCreated DATETIME NOT NULL,
    Creator INT,
    LastUpdated DATETIME,
    PRIMARY KEY (ID)
);
```

Messages: Represents a user created chat message
```
CREATE TABLE IF NOT EXISTS Messages (
    ID INT NOT NULL AUTO_INCREMENT,
    ChannelID INT NOT NULL,
    Body VARCHAR(255) NOT NULL,
    TimeCreated DATETIME NOT NULL,
    Creator INT NOT NULL,
    LastUpdated DATETIME,
    PRIMARY KEY (ID)
);
```
