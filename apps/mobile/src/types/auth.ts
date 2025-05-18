/**
 * Authentication-related types for the Tribe app
 */

export interface User {
  id: string;
  email: string;
  displayName: string;
}

export interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  isLoading: boolean;
}

export interface AuthContextType extends AuthState {
  login: (email: string) => Promise<void>;
  logout: () => Promise<void>;
} 