/**
 * Format a date string in a user-friendly format
 * @param {string} dateString - ISO date string
 * @returns {string} - Formatted date string
 */
export const formatDateString = (dateString) => {
  if (!dateString) return 'N/A';
  
  const date = new Date(dateString);
  
  // Check if date is valid
  if (isNaN(date.getTime())) {
    return 'Invalid date';
  }
  
  return date.toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
}; 