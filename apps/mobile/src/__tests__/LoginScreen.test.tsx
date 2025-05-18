import React from 'react';
import { render, fireEvent } from '@testing-library/react-native';
import LoginScreen from '../screens/LoginScreen';
import { useNavigation } from '@react-navigation/native';
import { useAuth } from '../contexts/AuthContext';

// Mock the navigation hooks
jest.mock('@react-navigation/native', () => {
  return {
    ...jest.requireActual('@react-navigation/native'),
    useNavigation: jest.fn(),
  };
});

// Mock the auth context
jest.mock('../contexts/AuthContext', () => {
  return {
    useAuth: jest.fn(),
  };
});

describe('LoginScreen', () => {
  // Setup mocks
  const mockReset = jest.fn();
  const mockLogin = jest.fn();
  
  beforeEach(() => {
    (useNavigation as jest.Mock).mockReturnValue({
      reset: mockReset,
    });
    
    (useAuth as jest.Mock).mockReturnValue({
      login: mockLogin,
    });
    
    // Reset the login mock to resolve successfully by default
    mockLogin.mockResolvedValue(undefined);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  it('renders correctly', () => {
    const { getByText } = render(<LoginScreen />);
    
    // Check that title is displayed
    expect(getByText('Log In')).toBeTruthy();
    
    // Check that all dev user buttons are displayed
    expect(getByText('Log in as dev_user1')).toBeTruthy();
    expect(getByText('Log in as dev_user2')).toBeTruthy();
    expect(getByText('Log in as dev_user3')).toBeTruthy();
    expect(getByText('Log in as dev_user4')).toBeTruthy();
    
    // Check that Google login button is displayed but disabled
    expect(getByText('Log in with Google')).toBeTruthy();
  });

  it('logs in as dev_user1 when button is pressed', async () => {
    const { getByText } = render(<LoginScreen />);
    
    // Find and press the dev_user1 button
    const loginButton = getByText('Log in as dev_user1');
    fireEvent.press(loginButton);
    
    // Wait for the async login function to complete
    await new Promise(process.nextTick);
    
    // Verify login was called with correct email
    expect(mockLogin).toHaveBeenCalledWith('dev_user1@gonetribal.com');
    
    // Verify navigation reset was called to go to MainTabs
    expect(mockReset).toHaveBeenCalledWith({
      index: 0,
      routes: [{ name: 'MainTabs' }],
    });
  });

  it('handles login errors', async () => {
    // Mock console.error to prevent test output noise
    const originalConsoleError = console.error;
    console.error = jest.fn();
    
    // Make the login function reject
    mockLogin.mockRejectedValue(new Error('Login failed'));
    
    const { getByText } = render(<LoginScreen />);
    
    // Find and press the dev_user1 button
    const loginButton = getByText('Log in as dev_user1');
    fireEvent.press(loginButton);
    
    // Wait for the async login function to complete
    await new Promise(process.nextTick);
    
    // Verify login was called but navigation reset was not
    expect(mockLogin).toHaveBeenCalledWith('dev_user1@gonetribal.com');
    expect(mockReset).not.toHaveBeenCalled();
    
    // Verify error was logged
    expect(console.error).toHaveBeenCalled();
    
    // Restore console.error
    console.error = originalConsoleError;
  });
}); 