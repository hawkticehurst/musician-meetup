(function () {
    "use strict";
  
    // Remember to always run the main.go file on port 4000 (vs the default port 80)
    // const BASE_URL = "http://localhost:4000/v1/summary";
  
    const BASE_URL = "https://api.info441summary.me/v1/sessions";
  
    /**
     *  Functions that will be called once the window is loaded
     *  Submit button will get click event listener and call fetchUrlSummary
     */
    window.addEventListener("load", () => {
        const button = id('signup');
        button.addEventListener('click', function(event){
          event.preventDefault();
          window.location = "signup.html";
        });
    });

    window.addEventListener("load", () => {
      const button = id('submit');
      button.addEventListener('click', function(event){
        event.preventDefault();
        authenticate();
      });
    });

    const authenticate = () => {
        const user = {
          email: idValue('Email'),
          password: idValue('Password'),
        }
        fetch(BASE_URL, {
          method: 'POST', // *GET, POST, PUT, DELETE, etc.
          headers: {
            'Content-Type': 'application/json'
            // 'Content-Type': 'application/x-www-form-urlencoded',
          },
          body: JSON.stringify(user) // body data type must match "Content-Type" header
        }).then(checkStatus)
        .catch(displayError)
    }

  /**
   * Function to handle the result of an unsuccessful fetch call
   * @param {Object} error - Error resulting from unsuccesful fetch call 
   */
  const displayError = (error) => {
    // Retrieve container for displaying error
    const metaContainer = id('meta-container');
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
    console.log("checkstatus");
    if (response.status === 401) {
      return Promise.reject(new Error("Invalid fields"));
    } else {
      window.location = "home.html";
    }
  }

})();