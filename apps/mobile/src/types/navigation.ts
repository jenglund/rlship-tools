/**
 * Type definitions for navigation in the Tribe app
 */

export type RootStackParamList = {
  Splash: undefined;
  Login: undefined;
  MainTabs: undefined;
  TribesStack: undefined;
  TribeDetails: { tribeId: string };
};

export type TabParamList = {
  Home: undefined;
  Tribes: undefined;
  Lists: undefined;
  Interest: undefined;
  Profile: undefined;
};

export type TribesStackParamList = {
  Tribes: undefined;
  TribeDetails: { tribeId: string };
}; 