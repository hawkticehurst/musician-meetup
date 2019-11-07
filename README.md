# INFO 441 Final Project Proposal

## Project Description

Our target audience is people who want to play music in a group. We want to build a platform in which people can find others to form music ensembles and enjoy playing music together. 

It can be hard for people playing an instrument to find people playing similar instruments, or an ensemble/band may be looking for a certain player to join their group (Ex: Jazz band needs a bass player). This platform will allow people to post public meetup events and message other players. This platform will also allow users to learn from other players nearby to them.

As developers, we are intrigued by the challenge that live-messaging with web sockets presents. We are also interested to see how connecting users chatting with actual meetup capabilities will go, how seamless we can make the transition.

## Technical Description

### Infrastructure

Users will interact exclusively with our website/domain container, hosted by AWS, and that (Docker) container will interact with our web server handlers. In our backend, user/meetup information will be stored in a MySQL database, and session information will be stored in a Redis database. Between our Web UI and our session information wel will have a Gateway setup, and this will interact with the Web UI using a REST API. 

Being created on Draw.io

| **Priority** | **User** | **Description** |
|--------------|----------|-----------------|
| P0 (High) | As a player | I want to message other music players to plan meetups and play music together |
| P0 (High) | As a player | I want to create public meetup events that other music players can discover to organize group ensembles |
| P0 (High) | As a user | I want to be able to create a user account |
| P1 (Med) | As a user | I want to be able to share music/tabs with another player |
| P2 (Low) | As a user | I want to have a calendar of my future events to organize all my music events |
| P2 (Low) | As a user | I want to post public performances so that other users can come and watch |


**For each of your user stories, describe in 2-3 sentences what your technical implementation strategy is. Explicitly note in bold which technology you are using (if applicable):**

We will be using **web sockets** in order to allow users to message one another.

We will be using a web microservice to handle the creation of meetup events, this microservice will run our gateway handler and our MySQL database, which will store meetup information.

We will be using a redis session store to allow users to login/logout. This will go from our front-end in html/css to our auth handler which is in Go.

We will be 



**Include a list available endpoints your application will provide and what is the purpose it serves. Ex. GET /driver/{id}**

API Endpoints

/v1/meetups/    (Homepage)
* GET: Respond with a cardview of public meetups that users want to organize, only viewable if signed in
- 200: Retrieve and return all meetups information
- 401: No sessionID or user not logged in
- 500: Server error

/v1/meetups/{meetupID}    (Meetup page)
* GET: Respond with a struct with info on the given meetup
- 200: Retrieve and return all the meetup's information
- 401: No sessionID or user not logged in
- 500: Server error
* DELETE: delete meetup event, only user who created meetup event can delete meetup event
- 200: Retrieve and return all meetups information
- 401: No sessionID or user not logged in
- 500: Server error

/v1/user/signin
* POST: gets the sign in page
- 200: Authenticate and log user in
- 401: Cannot sign in given credentials
- 500: Server error
* DELETE: sign user out
- 200: Successfully sign out user and let the user know with message
- 401: Session error, potentially no one logged in or distorted session
- 500: Server error

/v1/user/signup
* POST: post information to user database to create user account
- 200: Retrieve new user information
- 401: Could not create user, or invalid session
- 500: Server error

/v1/meetups/newmeetup
* POST: post new meetup event to meetup database associated with user
- 200: Post info to MySQL database, return meetup id
- 401: Invalid session
- 500: Server error

/v1/user/chat/
* GET: get chat page 
- 200: Retrieve and return chat page
- 401: No sessionID or user not logged in
- 500: Server error

### Database Schemas

We will use MySQL as our persistent database.

user: User represents a person who can log-in, message, and be a part of meetups on our site
```
{
    'email': 'user_email',
    'passwordhash': 'password_hash',
    'username': 'username',
    'datesignedup': 'date'

}
```

credentials: Represents what a user will input to log in to their account
```
{
    'email': 'user_email',
    'password': 'password'

}
```

meetup: Represents a proposed or completed meetup between multiple users at a certain location for a certain activity
```
{
    'userlist': [
        'user1'
    ],
    'location': 'address',
    'creator': 'username',
    'date': 'date',
    'starttime': 'starttime',
    'endtime': 'endtime'
}
```

