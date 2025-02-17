
// Import the Keycloak library to handle authentication.
import Keycloak from "keycloak-js";

// Create a new Keycloak instance with the configuration object.
const _kc = new Keycloak({
    // url: process.env.REACT_APP_KEYCLOAK_URL,
    url: "http://localhost:8180",
    realm: process.env.REACT_APP_KEYCLOAK_REALM,
    clientId: process.env.REACT_APP_KEYCLOAK_CLIENT_ID,
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
        onLoad: 'login-required',
        pkceMethod: 'S256',
        checkLoginIframe: false,
        // Token timeouts in seconds
        timeSkew: 0,
        tokenMinValidity: 30, // Start refreshing 30 seconds before expiry
        refreshToken: true,
        // Token lifetimes
        sessionTimeOutInSeconds: 1800, // 30 minutes
        refreshTokenTimeoutInSeconds: 259200, // 3 days
        // Silent refresh
        silentCheckSsoRedirectUri: window.location.origin + '/silent-check-sso.html',
        enableLogging: true,
        // Token refresh interval
        refreshTokenPeriod: 60 // Refresh token every minute if needed
    })
        .then((authenticated) => {
            if (!authenticated) {
                console.log("User is not authenticated..!");
            } else {
                console.log("User is authenticated");
                // Set up token refresh
                setInterval(() => {
                    _kc.updateToken(70).catch(() => {
                        console.log('Failed to refresh token');
                    });
                }, 60000); // Check every minute

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
