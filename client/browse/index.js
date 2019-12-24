(function () {
  "use strict";

  const BASE_URL = "https://api.info441summary.me/v1/events";
  const LOGOUT_URL = "https://api.info441summary.me/v1/sessions/mine";
  const JOINEVENT_URL = "https://api.info441summary.me/v1/events/join";

  /**
   * Functions that will be called once the window is loaded
   */
  window.addEventListener("load", function () {
    const logoutBtn = id('logout-btn')
    logoutBtn.addEventListener('click', logUserOut);

    const createEventBtn = id('submit');
    createEventBtn.addEventListener('click', createEvent);

    getEvents();
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
      .catch(displayErrorHomePage)

    deleteCookie("auth");
  }

  /**
   * createEvent makes a request to create a new meetup event
   */
  const createEvent = () => {
    const newEvent = {
      title: id('Title').value,
      datetime: id('DateTime').value,
      location: id('Location').value,
      description: id('Description').value,
    }

    fetch(BASE_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': getAuthToken()
      },
      body: JSON.stringify(newEvent)
    }).then(checkStatus)
      .then(getEvents)
      .catch(displayErrorForm)
  }

  /**
   * getEvents makes a request to retrieve all meetup events
   */
  function getEvents() {
    fetch(BASE_URL, {
      method: 'GET',
      headers: {
        'Authorization': getAuthToken()
      }
    })
      .then(checkStatus)
      .then(response => response.json())
      .then(displayCards)
      .catch(displayErrorHomePage)
  }

  /**
   * displayCards renders all meetup events in the given JSON object on the browse webpage
   * @param {JSON object} events a list of meetup events
   */
  async function displayCards(events) {
    for (let i = 0; i < events.length; i++) {
      const event = events[i];
      const joinBtn = document.createElement("button");
      const isMember = await userIsMember(event.id);
      if (isMember) {
        joinBtn.disabled = true;
        joinBtn.innerText = "You are a member of this event";
        joinBtn.setAttribute("type", "button");
        joinBtn.classList.add("btn");
        joinBtn.classList.add("btn-success");
        joinBtn.classList.add("join-btn");
      } else {
        joinBtn.innerText = "Join Event";
        joinBtn.setAttribute("type", "button");
        joinBtn.classList.add("btn");
        joinBtn.classList.add("btn-primary");
        joinBtn.classList.add("join-btn");
        joinBtn.addEventListener("click", function () {
          joinBtn.disabled = true;
          joinEvent(event.id);
        });
      }
      const card = document.createElement('div');
      card.className = 'card';

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
      card.appendChild(joinBtn);
      id("events-container").appendChild(card);
    }
  }

  /**
   * joinEvent makes a request to add the current user to the given meetup event
   * @param {string} id an id associated with a specific meetup event
   */
  function joinEvent(id) {
    const eventID = {
      id: id
    }

    fetch(JOINEVENT_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': getAuthToken()
      },
      body: JSON.stringify(eventID)
    }).then(checkStatus)
      .catch(displayErrorHomePage)
  }

  /**
   * displayErrorForm handles the result of an unsuccessful fetch call to 
   * create a new meetup event
   * @param {string} error an error message
   */
  const displayErrorForm = (error) => {
    const metaContainer = id('formError');
    if (metaContainer.classList.contains("hidden")) {
      metaContainer.classList.remove("hidden");
    }
    metaContainer.innerHTML = "";

    const errorMsg = document.createElement('h2');
    errorMsg.classList.add("error-msg");
    errorMsg.textContent = error;
    metaContainer.appendChild(errorMsg);
  }

  /**
   * displayErrorHomePage handles the result of an unsuccessful fetch call to 
   * log out the user, get all meetup events, or join a meetup event
   * @param {string} error an error message
   */
  const displayErrorHomePage = (error) => {
    const metaContainer = id('errorHome');
    if (metaContainer.classList.contains("hidden")) {
      metaContainer.classList.remove("hidden");
    }
    metaContainer.innerHTML = "";

    const errorMsg = document.createElement('h2');
    errorMsg.classList.add("error-msg");
    errorMsg.textContent = error;
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
   * userIsMember checks if the current user is part of a given meetup event
   * @param {string} eventID an id associated with a specific meetup event
   * @returns {boolean} a boolean representing whether the user is a member of 
   * the given meetup event or not
   */
  const userIsMember = async (eventID) => {
    const response = await fetch(JOINEVENT_URL, {
      method: 'GET',
      headers: {
        'Authorization': getAuthToken()
      }
    });
    const result = await response.json();
    for (let i = 0; i < result.length; i++) {
      if (result[i].id == eventID) {
        return true;
      }
    }
    return false;
  }

})();