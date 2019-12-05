(function () {
  "use strict";

  const BASE_URL = "https://api.info441summary.me/v1/events";
  const LOGOUT_URL = "https://api.info441summary.me/v1/sessions/mine";
  const JOINEVENT_URL = "https://api.info441summary.me/v1/events/join";

  /**
   *  Functions that will be called once the window is loaded
   */
  window.addEventListener("load", () => {
    getEvents();

    const logoutBtn = id('log-out')
    logoutBtn.addEventListener('click', logUserOut);

    const createEventBtn = id('submit');
    createEventBtn.addEventListener('click', createEvent);
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
      .catch(displayErrorHomePage)

    deleteCookie("auth");
  }

  // deleteCookie given the name of a cookie, this function will delete that cookie
  const deleteCookie = (name) => {
    document.cookie = name + '=;expires=Thu, 01 Jan 1970 00:00:01 GMT;';
  };

  // createEvent makes a request to create a new event
  const createEvent = () => {
    const newEvent = {
      title: idValue('Title'),
      datetime: idValue('DateTime'),
      location: idValue('Location'),
      description: idValue('Description'),
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
      .catch(displayErrorHomePage)
  }

  const displayCards = async (info) => {
    for (var i = 0; i < info.length; i++) {
      let data = info[i];
      let joinBtn = document.createElement("button");
      let isMember = await userIsMember(data.id);
      if (isMember) {
        joinBtn.innerText = "You are a member of this event";
        joinBtn.setAttribute("type", "button");
        joinBtn.classList.add("btn");
        joinBtn.classList.add("btn-secondary");
      } else {
        joinBtn.innerText = "Join Event";
        joinBtn.setAttribute("type", "button");
        joinBtn.classList.add("btn");
        joinBtn.classList.add("btn-primary");
        joinBtn.classList.add("join-btn");
        joinBtn.addEventListener("click", function () {
          joinBtn.disabled = true;
          joinEvent(data.id);
        });
      }
      let card = document.createElement('div');
      card.className = 'card';

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
      card.appendChild(joinBtn);
      id("events-container").appendChild(card);
    }
  }

  const joinEvent = (id) => {
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
   * Function to handle the result of an unsuccessful fetch call
   * @param {Object} error - Error resulting from unsuccesful fetch call 
   */
  const displayErrorForm = (error) => {
    // Retrieve container for displaying error
    const metaContainer = id('formError');
    if (metaContainer.classList.contains("hidden")) {
      metaContainer.classList.remove("hidden");
    }
    metaContainer.innerHTML = "";

    // Render error
    const errorMsg = document.createElement('h2');
    errorMsg.classList.add("error-msg");
    errorMsg.textContent = error;
    metaContainer.appendChild(errorMsg);
  }

  const displayErrorHomePage = (error) => {
    // Retrieve container for displaying error
    const metaContainer = id('errorHome');
    if (metaContainer.classList.contains("hidden")) {
      metaContainer.classList.remove("hidden");
    }
    metaContainer.innerHTML = "";

    // Render error
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

  const idValue = (idName) => {
    return document.getElementById(idName).value;
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

  //Check if user is part of event
  const userIsMember = async (eventID) => {
    const response = await fetch(JOINEVENT_URL, {
      method: 'GET', // *GET, POST, PUT, DELETE, etc.
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