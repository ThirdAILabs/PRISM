// // Import the Keycloak library to handle authentication.
// import Keycloak from "keycloak-js";

// // Create a new Keycloak instance with the configuration object.
// const _kc = new Keycloak({
//     url: "http://localhost:8180",
//     realm: "prism-user",
//     clientId: "prism-user-login-client", // Must match Keycloak client ID exactly
//     redirectUri: window.location.origin
// });


// /**
//  * Initializes the Keycloak instance.
//  * This function calls _kc.init() to start the authentication flow.
//  *
//  * The configuration object passed to init() includes:
//  * - onLoad: 'check-sso' tells Keycloak to verify if the user already has an active session.
//  * - silentCheckSsoRedirectUri: Specifies where to perform the silent check for single sign-on.
//  * - pkceMethod: 'S256' enables PKCE for enhanced authentication security.
//  *
//  * If the initialization resolves, the callback 'onAuthenticatedCallback' is called.
//  * If the user is not authenticated, a message is logged to the console.
//  *
//  * @param onAuthenticatedCallback {Function} Callback function to execute after successful authentication.
//  */
// const initKeycloak = (onAuthenticatedCallback) => {
//     _kc.init({
//         onLoad: 'login-required',  // force login if not already authenticated
//         pkceMethod: 'S256',
//         checkLoginIframe: false // Essential for modern browsers
//     })
//         .then((authenticated) => {
//             if (!authenticated) {
//                 // Log if user is not authenticated.
//                 console.log("user is not authenticated..!");
//             } else {
//                 // If authenticated, call the provided callback for further actions.
//                 onAuthenticatedCallback();
//             }
//         })
//         .catch((err) => {
//             // Log any initialization errors to the console.
//             console.error("Keycloak initialization error:", err);
//         });
// };

// const doLogin = _kc.login;

// const doLogout = _kc.logout;

// const getToken = () => _kc.token;
// console.log("Token is", _kc.token);
// const getTokenParsed = () => _kc.tokenParsed;

// const isLoggedIn = () => !!_kc.token;

// const updateToken = (successCallback) =>
//     _kc.updateToken(5)
//         .then(successCallback)
//         .catch(doLogin);

// const getUsername = () => _kc.tokenParsed?.preferred_username;

// const hasRole = (roles) => roles.some((role) => _kc.hasRealmRole(role));

// const UserService = {
//     initKeycloak,
//     doLogin,
//     doLogout,
//     isLoggedIn,
//     getToken,
//     getTokenParsed,
//     updateToken,
//     getUsername,
//     hasRole,
// };

// export default UserService;

// Import the Keycloak library to handle authentication.
import Keycloak from "keycloak-js";

// Create a new Keycloak instance with the configuration object.
const _kc = new Keycloak({
    url: "http://localhost:8180",
    realm: "prism-user",
    clientId: "prism-user-login-client",
    redirectUri: window.location.origin
});

/**
 * Initializes the Keycloak instance.
 * This function awaits the resolution of _kc.init() and then logs the token 
 * after successful authentication. If the user is not authenticated, it logs a message accordingly.
 *
 * @param onAuthenticatedCallback {Function} Callback function to execute after successful authentication.
 */
const initKeycloak = (onAuthenticatedCallback) => {
    _kc.init({
        onLoad: 'login-required',  // force login if not already authenticated
        pkceMethod: 'S256',
        checkLoginIframe: false // Essential for modern browsers
    })
        .then((authenticated) => {
            if (!authenticated) {
                console.log("User is not authenticated..!");
            } else {
                console.log("User is authenticated");
                console.log("Token is", _kc.token); // Token is now available here
                onAuthenticatedCallback();
            }
        })
        .catch((err) => {
            console.error("Keycloak initialization error:", err);
        });
};

const doLogin = _kc.login;
const doLogout = _kc.logout;

const getToken = () => _kc.token;
const getTokenParsed = () => _kc.tokenParsed;
const isLoggedIn = () => !!_kc.token;
const updateToken = (successCallback) =>
    _kc.updateToken(5)
        .then(successCallback)
        .catch(doLogin);

const getUsername = () => _kc.tokenParsed?.preferred_username;
const hasRole = (roles) => roles.some((role) => _kc.hasRealmRole(role));

const UserService = {
    initKeycloak,
    doLogin,
    doLogout,
    isLoggedIn,
    getToken,
    getTokenParsed,
    updateToken,
    getUsername,
    hasRole,
};

export default UserService;
