import React from 'react';
import { render, fireEvent } from '@testing-library/react-native';
import SplashScreen from '../screens/SplashScreen';
import { useNavigation } from '@react-navigation/native';

// Mock the navigation hooks
jest.mock('@react-navigation/native', () => {
  return {
    ...jest.requireActual('@react-navigation/native'),
    useNavigation: jest.fn(),
  };
});

describe('SplashScreen', () => {
  // Setup navigation mock
  const mockNavigate = jest.fn();
  beforeEach(() => {
    (useNavigation as jest.Mock).mockReturnValue({
      navigate: mockNavigate,
    });
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  it('renders correctly', () => {
    const { getByText } = render(<SplashScreen />);
    
    // Check that "Tribe" title is displayed
    expect(getByText('Tribe')).toBeTruthy();
    
    // Check that "Log In" button is displayed
    expect(getByText('Log In')).toBeTruthy();
  });

  it('navigates to Login screen when Log In button is pressed', () => {
    const { getByText } = render(<SplashScreen />);
    
    // Find and press the Log In button
    const loginButton = getByText('Log In');
    fireEvent.press(loginButton);
    
    // Verify navigation was called with correct screen name
    expect(mockNavigate).toHaveBeenCalledWith('Login');
  });
}); 