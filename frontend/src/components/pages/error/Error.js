import { useNavigate, useSearchParams } from 'react-router-dom';
import '../../common/tools/button/button1.css';
import './Error.css';

export default function Error() {
  const [searchParams] = useSearchParams();
  const message = searchParams.get('message');
  const status = searchParams.get('status');
  const navigate = useNavigate();

  const errorMessages = {
    400: ['Bad Request', 'The server could not understand the request due to invalid syntax.'],
    401: ['Unauthorized', 'Please sign in to access this resource.'],
    403: ['Forbidden', 'You do not have permission to access the requested resource.'],
    404: ['Not Found', 'The requested resource could not be found on this server.'],
    408: ['Request Timeout', 'Request The server timed out waiting for the request.'],
    422: ['Unprocessable Entity', 'The request was well-formed but contained semantic errors.'],
    500: ['Internal Server Error', 'Something went wrong on our side.'],
    503: ['Service Unavailable', 'The server is currently unable to handle the request.'],
  };

  return (
    <div class="error-page d-flex align-items-center justify-content-center">
      <div class="error-container text-center p-4">
        <h1 class="error-code mb-0">{status}</h1>
        <h2 class="display-6 error-message mb-3">{errorMessages[status][0]}</h2>
        <p class="lead error-message mb-5">{errorMessages[status][1]}</p>
        <div class="d-flex justify-content-center gap-3">
          <button
            className="button"
            onClick={() => {
              navigate('/');
            }}
            style={{ width: '250px' }}
          >
            Return home
          </button>
        </div>
      </div>
    </div>
  );
}
