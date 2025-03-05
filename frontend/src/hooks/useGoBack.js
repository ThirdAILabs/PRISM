import { useCallback } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

const useGoBack = (fallback) => {
    const navigate = useNavigate();
    const location = useLocation();

    return useCallback(() => {
        if (location?.state?.canGoBack) {
            navigate(-1);
        } else {
            navigate(fallback, { replace: true });
        }
    }, [location, fallback]);
}

export default useGoBack;