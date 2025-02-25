export default function Error({ message }) {
  return (
    <div className="error">
      <p>{message || 'An unexpected error occurred.'}</p>
    </div>
  );
}
