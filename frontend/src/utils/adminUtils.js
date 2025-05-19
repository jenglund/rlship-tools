/**
 * Check if a user is an admin based on their email domain
 * @param {Object} user - The user object from AuthContext
 * @returns {boolean} - True if the user is an admin
 */
export const isAdmin = (user) => {
  if (!user || !user.email) {
    return false;
  }
  
  // Check if email ends with @gonetribal.com
  return user.email.endsWith('@gonetribal.com');
}; 