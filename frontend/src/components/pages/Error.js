import { useSearchParams } from 'react-router-dom';

export default function Error() {
  const [searchParams] = useSearchParams();
  const message = searchParams.get('message');

  return (
    <div className="error" style={{ padding: '20px', textAlign: 'center' }}>
      <h2>Error</h2>
      <p>{message || 'An unexpected error occurred.'}</p>
    </div>
  );
}