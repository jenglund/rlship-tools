// Mock data for tribes
const MOCK_TRIBES = [
  {
    id: 1,
    name: 'Family',
    description: 'Family activities and planning',
    memberCount: 4,
    createdAt: '2023-05-15'
  },
  {
    id: 2,
    name: 'Friends',
    description: 'Weekend plans with friends',
    memberCount: 6,
    createdAt: '2023-06-21'
  },
  {
    id: 3,
    name: 'Book Club',
    description: 'Monthly book discussions',
    memberCount: 8,
    createdAt: '2023-07-10'
  },
  {
    id: 4,
    name: 'Hiking Group',
    description: 'Outdoor adventures',
    memberCount: 5,
    createdAt: '2023-08-03'
  }
];

// Get all tribes for a user
export const getUserTribes = (userId) => {
  // In a real implementation, this would be a fetch call to the API
  return new Promise((resolve) => {
    // Simulate API delay
    setTimeout(() => {
      resolve(MOCK_TRIBES);
    }, 500);
  });
};

// Get a single tribe by ID
export const getTribeById = (tribeId) => {
  return new Promise((resolve, reject) => {
    // Simulate API delay
    setTimeout(() => {
      const tribe = MOCK_TRIBES.find(tribe => tribe.id === parseInt(tribeId));
      if (tribe) {
        resolve(tribe);
      } else {
        reject(new Error('Tribe not found'));
      }
    }, 500);
  });
}; 