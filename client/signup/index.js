(function () {
  "use strict";

  const BASE_URL = "https://api.info441summary.me/v1/users";

  /**
   * Functions that will be called once the window is loaded
   */
  window.addEventListener("load", () => {
    id('submit').addEventListener('click', function (event) {
      event.preventDefault();
      createUser();
    });
  });

  /**
   * createUser makes a request to create a new user
   */
  const createUser = () => {
    const newUser = {
      firstName: id('FirstName').value,
      lastName: id('LastName').value,
      email: id('Email').value,
      userName: id('UserName').value,
      password: id('Password').value,
      passwordConf: id('PasswordConf').value
    }

    fetch(BASE_URL, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(newUser)
    }).then(checkStatus)
      .then(redirectToLogIn)
      .catch(displayError)
  }

  /**
   * displayError handles the result of an unsuccessful fetch to create a new user
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
   * redirectToLogIn will redirect to the log in page upon successfully completing 
   * the fetch request to create a new user
   */
  const redirectToLogIn = () => {
    window.location = "../index.html";
  }

})();