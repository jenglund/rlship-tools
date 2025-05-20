import React from 'react';
import { 
  FaMapMarkerAlt, 
  FaFilm, 
  FaRunning, 
  FaUtensils, 
  FaShoppingCart, 
  FaBook, 
  FaList, 
  FaMusic, 
  FaGamepad 
} from 'react-icons/fa';

// List type definitions with icons, labels, and metadata fields
export const LIST_TYPES = {
  location: {
    icon: <FaMapMarkerAlt />,
    label: 'Location',
    color: '#e74c3c',
    description: 'Places to visit, restaurants, venues, etc.',
    metadataFields: [
      { name: 'address', label: 'Address', type: 'text' },
      { name: 'latitude', label: 'Latitude', type: 'number', step: '0.000001' },
      { name: 'longitude', label: 'Longitude', type: 'number', step: '0.000001' },
      { name: 'category', label: 'Category', type: 'select', options: [
        { value: 'restaurant', label: 'Restaurant' },
        { value: 'bar', label: 'Bar/Pub' },
        { value: 'cafe', label: 'Cafe' },
        { value: 'attraction', label: 'Attraction' },
        { value: 'park', label: 'Park' },
        { value: 'store', label: 'Store' },
        { value: 'other', label: 'Other' }
      ]}
    ]
  },
  media: {
    icon: <FaFilm />,
    label: 'Media',
    color: '#3498db',
    description: 'Movies, TV shows, books, music, etc.',
    metadataFields: [
      { name: 'mediaType', label: 'Media Type', type: 'select', options: [
        { value: 'movie', label: 'Movie' },
        { value: 'tv', label: 'TV Show' },
        { value: 'book', label: 'Book' },
        { value: 'music', label: 'Music' },
        { value: 'other', label: 'Other' }
      ]},
      { name: 'creator', label: 'Creator/Author', type: 'text' },
      { name: 'releaseYear', label: 'Release Year', type: 'number' },
      { name: 'genre', label: 'Genre', type: 'text' },
      { name: 'length', label: 'Length/Duration', type: 'text' }
    ]
  },
  activity: {
    icon: <FaRunning />,
    label: 'Activity',
    color: '#2ecc71',
    description: 'Things to do, games, experiences, etc.',
    metadataFields: [
      { name: 'duration', label: 'Duration (minutes)', type: 'number' },
      { name: 'participants', label: 'Participants', type: 'text' },
      { name: 'cost', label: 'Estimated Cost ($)', type: 'number', step: '0.01' },
      { name: 'intensity', label: 'Intensity', type: 'select', options: [
        { value: 'low', label: 'Low' },
        { value: 'medium', label: 'Medium' },
        { value: 'high', label: 'High' }
      ]},
      { name: 'category', label: 'Category', type: 'select', options: [
        { value: 'indoors', label: 'Indoors' },
        { value: 'outdoors', label: 'Outdoors' },
        { value: 'sports', label: 'Sports' },
        { value: 'games', label: 'Games' },
        { value: 'arts', label: 'Arts & Crafts' },
        { value: 'other', label: 'Other' }
      ]}
    ]
  },
  food: {
    icon: <FaUtensils />,
    label: 'Food',
    color: '#f39c12',
    description: 'Recipes, meal ideas, restaurants, etc.',
    metadataFields: [
      { name: 'cuisine', label: 'Cuisine', type: 'text' },
      { name: 'prepTime', label: 'Prep Time (minutes)', type: 'number' },
      { name: 'cookTime', label: 'Cook Time (minutes)', type: 'number' },
      { name: 'ingredients', label: 'Main Ingredients', type: 'text' },
      { name: 'difficulty', label: 'Difficulty', type: 'select', options: [
        { value: 'easy', label: 'Easy' },
        { value: 'medium', label: 'Medium' },
        { value: 'hard', label: 'Hard' }
      ]}
    ]
  },
  shopping: {
    icon: <FaShoppingCart />,
    label: 'Shopping',
    color: '#9b59b6',
    description: 'Items to buy, gift ideas, etc.',
    metadataFields: [
      { name: 'price', label: 'Estimated Price ($)', type: 'number', step: '0.01' },
      { name: 'store', label: 'Store/Website', type: 'text' },
      { name: 'category', label: 'Category', type: 'text' },
      { name: 'priority', label: 'Priority', type: 'select', options: [
        { value: 'low', label: 'Low' },
        { value: 'medium', label: 'Medium' },
        { value: 'high', label: 'High' }
      ]}
    ]
  },
  reading: {
    icon: <FaBook />,
    label: 'Reading',
    color: '#1abc9c',
    description: 'Books, articles, blogs, etc.',
    metadataFields: [
      { name: 'author', label: 'Author', type: 'text' },
      { name: 'genre', label: 'Genre', type: 'text' },
      { name: 'pages', label: 'Pages', type: 'number' },
      { name: 'publicationYear', label: 'Publication Year', type: 'number' }
    ]
  },
  general: {
    icon: <FaList />,
    label: 'General',
    color: '#7f8c8d',
    description: 'A general purpose list for anything',
    metadataFields: [
      { name: 'category', label: 'Category', type: 'text' },
      { name: 'priority', label: 'Priority', type: 'select', options: [
        { value: 'low', label: 'Low' },
        { value: 'medium', label: 'Medium' },
        { value: 'high', label: 'High' }
      ]}
    ]
  },
  music: {
    icon: <FaMusic />,
    label: 'Music',
    color: '#e67e22',
    description: 'Songs, albums, artists, playlists, etc.',
    metadataFields: [
      { name: 'artist', label: 'Artist', type: 'text' },
      { name: 'album', label: 'Album', type: 'text' },
      { name: 'genre', label: 'Genre', type: 'text' },
      { name: 'year', label: 'Release Year', type: 'number' },
      { name: 'duration', label: 'Duration', type: 'text' }
    ]
  },
  games: {
    icon: <FaGamepad />,
    label: 'Games',
    color: '#8e44ad',
    description: 'Video games, board games, etc.',
    metadataFields: [
      { name: 'platform', label: 'Platform', type: 'text' },
      { name: 'genre', label: 'Genre', type: 'text' },
      { name: 'releaseYear', label: 'Release Year', type: 'number' },
      { name: 'publisher', label: 'Publisher', type: 'text' },
      { name: 'playerCount', label: 'Number of Players', type: 'text' }
    ]
  }
};

// Helper function to get list type info
export const getListTypeInfo = (type) => {
  return LIST_TYPES[type] || LIST_TYPES.general;
};

// Get icon for a list type
export const getListTypeIcon = (type) => {
  const typeInfo = getListTypeInfo(type);
  return typeInfo.icon;
};

// Get label for a list type
export const getListTypeLabel = (type) => {
  const typeInfo = getListTypeInfo(type);
  return typeInfo.label;
};

// Get color for a list type
export const getListTypeColor = (type) => {
  const typeInfo = getListTypeInfo(type);
  return typeInfo.color;
};

// Get metadata fields for a list type
export const getListTypeMetadataFields = (type) => {
  const typeInfo = getListTypeInfo(type);
  return typeInfo.metadataFields || [];
};

// Get all available list types as options for dropdown
export const getListTypeOptions = () => {
  return Object.entries(LIST_TYPES).map(([value, info]) => ({
    value,
    label: info.label,
    icon: info.icon,
    description: info.description
  }));
};

// Render a type-specific icon with optional label
export const renderListTypeIcon = (type, includeLabel = false, className = '') => {
  const typeInfo = getListTypeInfo(type);
  return (
    <span className={`list-type-icon ${className}`} style={{ color: typeInfo.color }}>
      {typeInfo.icon} {includeLabel && typeInfo.label}
    </span>
  );
};

export default {
  LIST_TYPES,
  getListTypeInfo,
  getListTypeIcon,
  getListTypeLabel,
  getListTypeColor,
  getListTypeMetadataFields,
  getListTypeOptions,
  renderListTypeIcon
}; 