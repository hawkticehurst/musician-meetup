(function () {
  "use strict";

  // Remember to always run the main.go file on port 4000 (vs the default port 80)
  // const BASE_URL = "http://localhost:4000/v1/summary";

  const BASE_URL = "https://api.info441summary.me/v1/events";
  const LOGOUT_URL = "https://api.info441summary.me/v1/sessions/mine";
  const CHANNEL_URL = "https://api.info441summary.me/v1/channels/";
  let CURR_CHANNEL = undefined; // This is the currently opened channel id

  /**
   *  Functions that will be called once the window is loaded
   *  Submit button will get click event listener and call fetchUrlSummary
   */
  window.addEventListener("load", () => {
    getEvents();
    const logoutButton = id('logout-btn')
    logoutButton.addEventListener('click', function (event) {
      //event.preventDefault();
      logUserOut();
    });

    const sendBtn = id("send-btn");
    sendBtn.addEventListener("click", sendMessage);
  });

  const logUserOut = () => {
    fetch(LOGOUT_URL, {
      method: 'DELETE', // *GET, POST, PUT, DELETE, etc.
      headers: {
        'Content-Type': 'text/html',
        'Authorization': getAuthToken()
      },
      body: ""// body data type must match "Content-Type" header
    }).then(checkStatus)
      .then(window.location.pathname = '../').catch(displayErrorHomePage)
  }

  const getEvents = () => {
    fetch(BASE_URL, {
      method: 'GET', // *GET, POST, PUT, DELETE, etc.
      headers: {
        'Authorization': getAuthToken()
      }
    })
      .then(checkStatus)
      .then(response => response.json())
      .then(displayCards)
      .catch(displayErrorHomePage);
  }

  const displayCards = (info) => {
    for (var i = 0; i < info.length; i++) {
      let data = info[i];
      let card = document.createElement('div');
      card.className = 'card';

      card.addEventListener("click", function () {
        clearChannel();
        getChannel(data.channel);
        CURR_CHANNEL = data.channel;
      });

      let title = document.createElement('h4');
      title.innerText = data.title;
      title.className = 'card-title';

      let datetime = document.createElement('p');
      datetime.innerText = data.time;
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

  const clearChannel = () => {
    id("channel").innerHTML = "";
  }

  const getChannel = (channelID) => {
    console.log("getChannel() ran");
    //get all messages
    const channelURL = CHANNEL_URL + channelID;
    fetch(channelURL, {
      method: 'GET', // *GET, POST, PUT, DELETE, etc.
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
        method: 'POST', // *GET, POST, PUT, DELETE, etc.
        headers: {
          'Content-Type': 'application/json',
          'Authorization': getAuthToken()
        },
        body: JSON.stringify(messageJSON)
      })
        .then(checkStatus)
        .then(function (response) {
          console.log("Successful check status");
        })
        .then(function (response) {
          console.log("Inside getChan anonymous function");
          getChannel(CURR_CHANNEL);
        })
        .catch(displayErrorHomePage);
    }
  }

  const displayMessages = (data) => {
    for (var i = 0; i < data.length; i++) {
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
    // Retrieve container for displaying error
    const metaContainer = id('error-card');
    const cardsContainer = id("cards-container");
    if (metaContainer.classList.contains("hidden")) {
      metaContainer.classList.remove("hidden");
      cardsContainer.classList.add("hidden");
    }
    metaContainer.innerHTML = "";

    // Render error
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

  const idValue = (idName) => {
    return document.getElementById(idName).value;
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
    if (response.status >= 200 && response.status < 300) {
      return response;
    } else {
      return Promise.reject(new Error(response.status + ": " + response.statusText));
    }
  }

})();