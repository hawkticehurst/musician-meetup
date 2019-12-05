(function () {
  "use strict";

  const BASE_URL = "https://api.info441summary.me/v1/events/join";
  const LOGOUT_URL = "https://api.info441summary.me/v1/sessions/mine";
  const CHANNEL_URL = "https://api.info441summary.me/v1/channels/";
  let CURR_CHANNEL = undefined; // This is the currently opened chat channel id

  /**
   *  Functions that will be called once the window is loaded
   */
  window.addEventListener("load", () => {
    getEvents();

    const logoutBtn = id('logout-btn')
    logoutBtn.addEventListener('click', logUserOut);

    const sendBtn = id("send-btn");
    sendBtn.addEventListener("click", sendMessage);
  });

  // Establish WebSocket connection
  let sock;
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

        let messageBox = document.createElement('div');
        messageBox.className = 'container';

        let name = document.createElement('p');
        name.innerText = msgInfo.creator.firstName + " " + msgInfo.creator.lastName;

        let message = document.createElement('p');
        message.innerText = msgInfo.body;

        messageBox.appendChild(name);
        messageBox.appendChild(message);
        id("channel").appendChild(messageBox);
      }
    };
  });

  // logUserOut will make request to log out the current user and delete their auth cookie
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

  // deleteCookie given the name of a cookie, this function will delete that cookie
  const deleteCookie = (name) => {
    document.cookie = name + '=;expires=Thu, 01 Jan 1970 00:00:01 GMT;';
  };

  // getEvents fetches the data in order to display chat event cards
  const getEvents = () => {
    fetch(BASE_URL, {
      method: 'GET',
      headers: {
        'Authorization': getAuthToken()
      }
    })
      .then(checkStatus)
      .then(response => response.json())
      .then(displayCards)
      .catch(displayErrorHomePage);
  }

  // displayCards displays chat event cards
  const displayCards = (info) => {
    for (let i = 0; i < info.length; i++) {
      let data = info[i];
      let card = document.createElement('div');
      card.className = 'card';

      card.addEventListener("click", function () {
        clearChannel();
        getChannel(data.channel);
        resetChannelBgs();
        showTextBar();
        card.className = 'card bg-primary'
        CURR_CHANNEL = data.channel;
      });

      let title = document.createElement('h4');
      title.innerText = data.title;
      title.className = 'card-title';

      let datetime = document.createElement('p');
      datetime.innerText = data.datetime;
      datetime.className = 'card-text';

      let location = document.createElement('p');
      location.innerText = data.location;
      location.className = 'card-text';

      let description = document.createElement('p');
      description.innerText = data.description;
      description.className = 'card-text';

      card.appendChild(title);
      card.appendChild(datetime);
      card.appendChild(location);
      card.appendChild(description);
      id("cards-container").appendChild(card);
    }
  }

  const showTextBar = () => {
    let chatContainer = id("chat-input-container");
    chatContainer.classList.remove("d-none");

    let chatInput = id("chat-input");
    chatInput.classList.remove("d-none");

    let sendBtn = id("send-btn");
    sendBtn.classList.remove("d-none");
  }

  const resetChannelBgs = () => {
    let cards = id("cards-container").querySelectorAll(".card");

    for (let i = 0; i < cards.length; i++) {
      cards[i].className = 'card';
    }
  }

  const clearChannel = () => {
    id("channel").innerHTML = "";
  }

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

  const displayMessages = (data) => {
    for (let i = 0; i < data.length; i++) {
      let info = data[i];
      let messageBox = document.createElement('div');
      messageBox.className = 'container';

      let name = document.createElement('p');
      name.innerText = info.creator.firstName + " " + info.creator.lastName;

      let message = document.createElement('p');
      message.innerText = info.body;

      messageBox.appendChild(name);
      messageBox.appendChild(message);
      id("channel").appendChild(messageBox);
    }
  }

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

  const getAuthToken = () => {
    let nameEQ = "auth=";
    let cookies = document.cookie.split(";");
    for (let i = 0; i < cookies.length; i++) {
      let cookie = cookies[i];
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
   * Helper function to return the response's result text if successful, otherwise
   * returns the rejected Promise result with an error status and corresponding text
   * @param {Object} response Response to check for success/error
   * @returns {Object} Valid result text if response was successful, otherwise rejected
   *                   Promise result
   */
  const checkStatus = (response) => {
    //console.log("inside check status, status code:" + response.status)
    if (response.status >= 200 && response.status < 300) {
      return response;
    } else {
      return Promise.reject(new Error(response.status + ": " + response.statusText));
    }
  }

})();