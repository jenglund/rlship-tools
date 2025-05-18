import React from 'react';
import { render, waitFor } from '@testing-library/react-native';
import TribesScreen from '../screens/TribesScreen';
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

describe('TribesScreen', () => {
  // Setup mocks
  const mockNavigate = jest.fn();
  
  beforeEach(() => {
    (useNavigation as jest.Mock).mockReturnValue({
      navigate: mockNavigate,
    });
    
    (useAuth as jest.Mock).mockReturnValue({
      user: {
        id: 'dev_user1',
        email: 'dev_user1@gonetribal.com',
        displayName: 'Dev User 1',
      }
    });
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  it('renders loading state initially', () => {
    const { getByText } = render(<TribesScreen />);
    
    // Check that loading indicator is shown
    expect(getByText('Loading your tribes...')).toBeTruthy();
  });

  it('renders the list of tribes after loading', async () => {
    const { getByText, queryByText } = render(<TribesScreen />);
    
    // Wait for loading state to finish
    await waitFor(() => {
      expect(queryByText('Loading your tribes...')).toBeNull();
    });
    
    // Check that title is displayed
    expect(getByText('Your Tribes')).toBeTruthy();
    
    // Check that all expected tribes are rendered
    expect(getByText('Family')).toBeTruthy();
    expect(getByText('Friends')).toBeTruthy();
    expect(getByText('Work Buddies')).toBeTruthy();
    expect(getByText('Book Club')).toBeTruthy();
    
    // Check tribe descriptions
    expect(getByText('Family activities and gatherings')).toBeTruthy();
    expect(getByText('Weekend adventures with friends')).toBeTruthy();
    expect(getByText('Work team activities')).toBeTruthy();
    expect(getByText('Monthly book discussions')).toBeTruthy();
    
    // Check member counts
    expect(getByText('4 members')).toBeTruthy();
    expect(getByText('6 members')).toBeTruthy();
    expect(getByText('8 members')).toBeTruthy();
    expect(getByText('5 members')).toBeTruthy();
  });

  it('renders empty state when user has no tribes', async () => {
    // Mock the fetchUserTribes function to return an empty array
    jest.mock('../screens/TribesScreen', () => {
      const originalModule = jest.requireActual('../screens/TribesScreen');
      return {
        ...originalModule,
        fetchUserTribes: jest.fn().mockResolvedValue([]),
      };
    });

    const { getByText, queryByText } = render(<TribesScreen />);
    
    // Wait for loading state to finish
    await waitFor(() => {
      expect(queryByText('Loading your tribes...')).toBeNull();
    });
    
    // Check that empty state message is displayed
    expect(getByText('You aren\'t a member of any tribes yet.')).toBeTruthy();
  });

  it('navigates to tribe details when a tribe is pressed', async () => {
    const { getByText, queryByText } = render(<TribesScreen />);
    
    // Wait for loading state to finish
    await waitFor(() => {
      expect(queryByText('Loading your tribes...')).toBeNull();
    });
    
    // Find and press the Family tribe
    const familyTribeCard = getByText('Family');
    familyTribeCard.parent?.props.onPress();
    
    // Verify navigation was called with correct params
    expect(mockNavigate).toHaveBeenCalledWith('TribeDetails', { tribeId: 'tribe1' });
  });
}); 