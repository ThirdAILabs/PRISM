// import { Navigate, Outlet } from "react-router-dom";
// import UserService from "../services/userService";

// const PrivateComponent = () => {
//     const auth = UserService.isLoggedIn();
//     return auth ? <Outlet /> : <Navigate to={"/login"} />;
// };

// export default PrivateComponent;


// privateComponent.js
import { useEffect } from 'react';
import { Navigate, Outlet } from "react-router-dom";
import UserService from "../services/userService";

const PrivateComponent = () => {
    useEffect(() => {
        if (!UserService.isLoggedIn()) {
            UserService.doLogin(); // Force immediate redirect
        }
    }, []);

    return UserService.isLoggedIn() ? <Outlet /> : <Navigate to="/login" />;
};
