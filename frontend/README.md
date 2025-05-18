# Tribe Frontend

This directory contains the frontend web application for Tribe, built with React 19.

## Overview

Tribe is an application designed for relationship management and activity planning. It helps groups of people (tribes) share and plan activities together. While initially focused on romantic couples, the app is designed to be inclusive of various relationship types including polyamorous relationships, friend groups, roommates, families, etc.

## Technology Stack

- React 19
- Modern JavaScript (ES6+)
- Webpack for bundling
- Babel for transpilation

## Getting Started

### Prerequisites

- Node.js (v18+ recommended)
- npm

### Installation

```bash
# Install dependencies
npm install
```

### Development

```bash
# Start the development server
npm start
```

This will open the application in your default browser at http://localhost:3000.

### Building for Production

```bash
# Create a production build
npm run build
```

The build artifacts will be stored in the `dist` directory.

## Key Features

- Tribes - Groups of people sharing activities and plans
- Lists of places/activities to visit or do together
- Calendar integration for scheduling
- Support for various relationship types

## Project Structure

```
frontend/
├── public/           # Static files
├── src/              # Source code
│   ├── components/   # React components
│   ├── contexts/     # React contexts for state management
│   ├── hooks/        # Custom React hooks
│   ├── services/     # API services
│   ├── utils/        # Utility functions
│   ├── App.js        # Main App component
│   └── index.js      # Application entry point
└── webpack.config.js # Webpack configuration
``` 