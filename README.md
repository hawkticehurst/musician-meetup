# INFO 441 Final Project Proposal

## Project Description

This is a written, non-technical description of your project. Depending on the specifics of your project, you should outline the answers to these (and perhaps other) questions:

- Who is your target audience?  Who do you envision using your application? Depending on the domain of your application, there may be a variety of audiences interested in using your application.  You should hone in on one of these audiences.
- Why does your audience want to use your application? Please provide some sort of reasoning. 
- Why do you as developers want to build this application?

**Write this as a narrative (no bullet points or tables).**

Our target audience is people who want to play music in a group. We want to build a platform in which people can find others to form music ensembles to enjoy playing music. It can be hard for people playing an instrument to find people playing similar instruments, or an ensemble/band may be looking for a certain player to join their group (Ex: Jazz band needs a bass player). This platform will allow people to post public meetup events and message other players.

## Technical Description

**Include an architectural diagram mapping 1) all server and database components, 2) flow of data, and its communication type (message? REST API?).**

### Infrastructure

Users will interact exclusively with our website/domain container, hosted by AWS, and that (Docker) container will interact with our web server handlers. In our backend, user/meetup information will be stored in a MySQL database, and session information will be stored in a Redis database. Between our Web UI and our session information wel will have a Gateway setup, and this will interact with the Web UI using a REST API. 

Being created on Draw.io

**A summary table of user stories (descriptions of the experience a user would want) in the following format (P0 P1 means priority. These classifications do not factor into grading. They are more for your own benefit to think about what would be top priorities if you happened to run out of time. Mark which ones you think make up the minimal viable product)**

| **Priority** | **User** | **Description** |
|--------------|----------|-----------------|
| P0 (High) | As a user | I want to message other music players to plan meetups and play music together |
| P0 (High) | As a user | I want to create public meetup events that other music players can discover to organize group ensembles |
| P0 (High) | As a user | I want to be able to create a user account |
| P1 (Med) | As a user | I want to add information on the instruments that I play, experience level, and music genre interests to my profile so that users can find players more easily to play with |
| P2 (Low) | As a user | I want to be able to share music/tabs with another player |
| P2 (Low) | As a user | I want to have a calendar of my future events to organize all my music events |
| P2 (Low) | As a user | I want to post public performances so that other users can come and watch |


**For each of your user stories, describe in 2-3 sentences what your technical implementation strategy is. Explicitly note in bold which technology you are using (if applicable):**

We will be using **web sockets** in order to allow users to message one another. 

**Include a list available endpoints your application will provide and what is the purpose it serves. Ex. GET /driver/{id}**

API Endpoints

/v1/meetups/    (Homepage)
* GET: Respond with a cardview of public meetups that users want to organize, only viewable if signed in
- 200: Retrieve and return all meetups information
- 401: No sessionID or user not logged in
- 500: Server error

/v1/meetups/    (Homepage)
* GET: Respond with a cardview of public meetups that users want to organize, only viewable if signed in
- 200: Retrieve and return all meetups information
- 401: No sessionID or user not logged in
- 500: Server error

/v1/user/signin
* GET: gets the sign in page
- 200: Retrieve and return all meetups information
- 401: No sessionID or user not logged in
- 500: Server error
* POST: post information to user database for authentication
- 201: Retrieve and return all user information, user struct
- 401: No sessionID or user not logged in
- 500: Server error

/v1/user/signup
* GET: gets the sign up page
- 200: Retrieve and return all meetups information
- 401: No sessionID or user not logged in
- 500: Server error
* POST: post information to user database to create user account
- 200: Retrieve and return all meetups information
- 401: No sessionID or user not logged in
- 500: Server error

/v1/meetups/{userID}
* POST: post new meetup event to meetup database associated with user
- 200: Retrieve and return all meetups information
- 401: No sessionID or user not logged in
- 500: Server error
* DELETE: delete meetup event, only user who created meetup event can delete meetup event
- 200: Retrieve and return all meetups information
- 401: No sessionID or user not logged in
- 500: Server error

/v1/user/chat/
* GET: get chat page 
- 200: Retrieve and return all meetups information
- 401: No sessionID or user not logged in
- 500: Server error

**Include any database schemas as an appendix**

### Database Schemas

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

