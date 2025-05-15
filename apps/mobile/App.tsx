import React from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { Provider as PaperProvider } from 'react-native-paper';
import { StatusBar } from 'expo-status-bar';

// Import screens (we'll create these next)
import HomeScreen from './src/screens/HomeScreen';
import ListsScreen from './src/screens/ListsScreen';
import InterestButtonsScreen from './src/screens/InterestButtonsScreen';
import ProfileScreen from './src/screens/ProfileScreen';

const Tab = createBottomTabNavigator();
const Stack = createNativeStackNavigator();

function TabNavigator() {
  return (
    <Tab.Navigator>
      <Tab.Screen 
        name="Home" 
        component={HomeScreen}
        options={{
          title: 'Home'
        }}
      />
      <Tab.Screen 
        name="Lists" 
        component={ListsScreen}
        options={{
          title: 'Lists'
        }}
      />
      <Tab.Screen 
        name="Interest" 
        component={InterestButtonsScreen}
        options={{
          title: 'Interest Buttons'
        }}
      />
      <Tab.Screen 
        name="Profile" 
        component={ProfileScreen}
        options={{
          title: 'Profile'
        }}
      />
    </Tab.Navigator>
  );
}

export default function App() {
  return (
    <PaperProvider>
      <NavigationContainer>
        <Stack.Navigator>
          <Stack.Screen
            name="MainTabs"
            component={TabNavigator}
            options={{ headerShown: false }}
          />
        </Stack.Navigator>
        <StatusBar style="auto" />
      </NavigationContainer>
    </PaperProvider>
  );
}
