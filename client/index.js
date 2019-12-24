(function () {
  "use strict";

  const BASE_URL = "https://api.info441summary.me/v1/sessions";

  /**
   * Functions that will be called once the window is loaded
   */
  window.addEventListener("load", () => {
    id('submit').addEventListener('click', function (event) {
      event.preventDefault();
      authenticateUser();
    });
  });

  /**
   * authenticateUser makes a request to authenticate (i.e. log in) the current user
   */
  const authenticateUser = () => {
    const user = {
      email: id('Email').value,
      password: id('Password').value,
    }

    fetch(BASE_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(user)
    }).then(checkStatus)
      .then(setAuthCookie)
      .then(redirectToBrowse)
      .catch(displayError)
  }

  /**
   * displayError handles the result of an unsuccessful fetch to authenticate 
   * (i.e. log in) the current user
   * @param {string} error an error message
   */
  const displayError = (error) => {
    const metaContainer = id('meta-container');
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
   * setAuthCookie sets an authorization cookie for the current user
   * @param {Response Object} response an HTTP response object
   */
  const setAuthCookie = (response) => {
    document.cookie = "auth=" + response.headers.get("authorization");
  }

  /**
   * redirectToBrowse will redirect to the browse page upon successfully completing
   * the fetch request to authenticate (i.e. log in) the user
   */
  const redirectToBrowse = () => {
    window.location = "./browse/index.html";
  }

})();