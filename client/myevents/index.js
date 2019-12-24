(function () {
  "use strict";

  const BASE_URL = "https://api.info441summary.me/v1/events/join";
  const LOGOUT_URL = "https://api.info441summary.me/v1/sessions/mine";
  const CHANNEL_URL = "https://api.info441summary.me/v1/channels/";
  let CURR_CHANNEL = undefined; // The id of the currently opened chat
  let sock; // A WebSocket connection

  /**
   * Functions that will be called once the window is loaded
   */
  window.addEventListener("load", () => {
    const logoutBtn = id('logout-btn')
    logoutBtn.addEventListener('click', logUserOut);

    const sendBtn = id("send-btn");
    sendBtn.addEventListener("click", sendMessage);

    getEvents();
  });

  /**
   * Establish a WebSocket connection
   */
  document.addEventListener("DOMContentLoaded", (event) => {
    sock = new WebSocket("wss://api.info441summary.me/v1/ws?auth=" + getAuthToken());
    sock.onopen = () => {
      console.log("Connection Opened");
    };
    sock.onclose = () => {
      console.log("Connection Closed");
    };
    sock.onmessage = (msg) => {
      let msgInfo = "";
      try {
        msgInfo = JSON.parse(msg.data).message;
      } catch (e) {
        console.log(e);
      }

      if (msgInfo.channelID == CURR_CHANNEL) {
        const messageContainer = document.createElement('div');
        messageContainer.className = 'container';
        const name = document.createElement('p');
        name.textContent = msgInfo.creator.firstName + " " + msgInfo.creator.lastName;
        const message = document.createElement('p');
        message.textContent = msgInfo.body;

        messageContainer.appendChild(name);
        messageContainer.appendChild(message);
        id("channel").appendChild(messageContainer);
      }
    };
  });

  /**
   * logUserOut will make request to log out the current user and delete their auth cookie
   */
  const logUserOut = () => {
    fetch(LOGOUT_URL, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'text/html',
        'Authorization': getAuthToken()
      },
      body: ""
    }).then(checkStatus)
      .then(window.location.pathname = '../')
      .catch(displayErrorHomePage);

    deleteCookie("auth");
  }

  /**
   * sendMessage sends a message to the currently opened chat
   */
  const sendMessage = () => {
    const message = id("chat-input").value;
    if (message.length > 0) {
      id("chat-input").value = "";

      const messageJSON = {
        "body": message
      }

      const url = CHANNEL_URL + CURR_CHANNEL;
      fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': getAuthToken()
        },
        body: JSON.stringify(messageJSON)
      })
        .then(checkStatus)
        .catch(displayErrorHomePage);
    }
  }

  /**
   * getEvents fetches all the meetup events the current user is a member of
   */
  const getEvents = () => {
    fetch(BASE_URL, {
      method: 'GET',
      headers: {
        'Authorization': getAuthToken()
      }
    })
      .then(checkStatus)
      .then(response => response.json())
      .then(displayEvents)
      .catch(displayErrorHomePage);
  }

  /**
   * displayEvents renders all meetup event chat cards contained in the given events object
   * @param {JSON Object} events a list of meetup events the current user is a member of
   */
  const displayEvents = (events) => {
    for (let i = 0; i < events.length; i++) {
      const event = events[i];
      const card = document.createElement('div');
      card.className = 'card';

      card.addEventListener("click", function () {
        clearChannel();
        getChannel(event.channel);
        resetChannelBackgrounds();
        showChatInput();
        card.className = 'card bg-primary'
        CURR_CHANNEL = event.channel;
      });

      const title = document.createElement('h4');
      title.innerText = event.title;
      title.className = 'card-title';

      const datetime = document.createElement('p');
      datetime.innerText = event.datetime;
      datetime.className = 'card-text';

      const location = document.createElement('p');
      location.innerText = event.location;
      location.className = 'card-text';

      const description = document.createElement('p');
      description.innerText = event.description;
      description.className = 'card-text';

      card.appendChild(title);
      card.appendChild(datetime);
      card.appendChild(location);
      card.appendChild(description);
      id("cards-container").appendChild(card);
    }
  }

  /**
   * getChannel fetches all the chat messages for the given chat channel
   * @param {string} channelID an id associated with a specific chat channel
   */
  const getChannel = (channelID) => {
    const channelURL = CHANNEL_URL + channelID;
    fetch(channelURL, {
      method: 'GET',
      headers: {
        'Authorization': getAuthToken()
      }
    })
      .then(checkStatus)
      .then(response => response.json())
      .then(displayMessages)
      .catch(displayErrorHomePage);
  }

  /**
   * displayMessages renders all the messages contained in the given messages object
   * @param {JSON Object} messages a list of chat messages
   */
  const displayMessages = (messages) => {
    for (let i = 0; i < messages.length; i++) {
      const msg = messages[i];
      const messageBox = document.createElement('div');
      messageBox.className = 'container';
      const name = document.createElement('p');
      name.innerText = msg.creator.firstName + " " + msg.creator.lastName;
      const message = document.createElement('p');
      message.innerText = msg.body;

      messageBox.appendChild(name);
      messageBox.appendChild(message);
      id("channel").appendChild(messageBox);
    }
  }

  /**
   * displayErrorHomePage handles the result of an unsuccessful fetch call to log out the 
   * user, send a chat message, get all meetup events the user is a member of, or get all
   * the chat messages of a given chat channel
   * @param {string} error an error message
   */
  const displayErrorHomePage = (error) => {
    const metaContainer = id('error-card');
    const cardsContainer = id("cards-container");
    if (metaContainer.classList.contains("hidden")) {
      metaContainer.classList.remove("hidden");
      cardsContainer.classList.add("hidden");
    }
    metaContainer.innerHTML = "";

    const errorMsg = document.createElement('h3');
    errorMsg.classList.add("error-msg");
    errorMsg.textContent = "Sorry we are unable to retrieve events at this time.";
    metaContainer.appendChild(errorMsg);
  }


  /* ------------------------------ Helper Functions  ------------------------------ */

  /**
   * Returns the element that has the ID attribute with the specified value.
   * @param {String} idName HTML element ID.
   * @returns {Object} DOM object associated with ID.
   */
  const id = (idName) => {
    return document.getElementById(idName);
  }

  /**
   * Helper function to return the response's result text if successful, otherwise
   * returns the rejected Promise result with an error status and corresponding text
   * @param {Object} response Response to check for success/error
   * @returns {Object} Valid result text if response was successful, otherwise rejected
   *                   Promise result
   */
  const checkStatus = (response) => {
    if (response.status >= 200 && response.status < 300) {
      return response;
    } else {
      return Promise.reject(new Error(response.status + ": " + response.statusText));
    }
  }

  /**
   * deleteCookie deletes a given cookie
   * @param {string} cookie a cookie
   */
  const deleteCookie = (cookie) => {
    document.cookie = cookie + '=;expires=Thu, 01 Jan 1970 00:00:01 GMT;';
  };

  /**
   * getAuthToken returns the authentication token of the given user or null if the token 
   * does not exist
   * @return {string, null} authentication token of the given user or null if the token 
   * does not exist
   */
  const getAuthToken = () => {
    const nameEQ = "auth=";
    const cookies = document.cookie.split(";");
    for (let i = 0; i < cookies.length; i++) {
      const cookie = cookies[i];
      while (cookie.charAt(0) == " ") {
        cookie = cookie.substring(1, cookie.length);
      }
      if (cookie.indexOf(nameEQ) == 0) {
        return cookie.substring(nameEQ.length, cookie.length);
      }
    }
    return null;
  }

  /**
   * clearChannel clears the chat container of all the currently rendered chat messages
   */
  const clearChannel = () => {
    id("channel").innerHTML = "";
  }

  /**
   * resetChannelBackgrounds resets the background of all chat channel cards 
   */
  const resetChannelBackgrounds = () => {
    const cards = id("cards-container").querySelectorAll(".card");
    for (let i = 0; i < cards.length; i++) {
      cards[i].className = 'card';
    }
  }

  /**
   * showChatInput renders the chat input and send button when a chat event card is clicked
   */
  const showChatInput = () => {
    const chatInputContainer = id("chat-input-container");
    chatInputContainer.classList.remove("hidden");
    const chatInput = id("chat-input");
    chatInput.classList.remove("hidden");
    const sendBtn = id("send-btn");
    sendBtn.classList.remove("hidden");
  }

})();