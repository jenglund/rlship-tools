import React, { useState } from 'react';
import { View, StyleSheet, ScrollView } from 'react-native';
import { Avatar, List, Button, Divider, Text } from 'react-native-paper';
import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { RootStackParamList } from '../types/navigation';
import { useAuth } from '../contexts/AuthContext';

type Relationship = {
  id: string;
  name: string;
  avatar: string;
};

export default function ProfileScreen() {
  const navigation = useNavigation<NativeStackNavigationProp<RootStackParamList>>();
  const { logout } = useAuth();
  
  // Example user data - in real app, this would come from your backend
  const [user] = useState({
    name: 'Alex Smith',
    email: 'alex@example.com',
    avatar: 'https://i.pravatar.cc/150?img=12',
  });

  // Example relationships - in real app, this would come from your backend
  const [relationships] = useState<Relationship[]>([
    { id: '1', name: 'Jordan Lee', avatar: 'https://i.pravatar.cc/150?img=13' },
    { id: '2', name: 'Sam Taylor', avatar: 'https://i.pravatar.cc/150?img=14' },
  ]);

  const handleLogout = async () => {
    try {
      await logout();
      navigation.reset({
        index: 0,
        routes: [{ name: 'Splash' }],
      });
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  return (
    <ScrollView style={styles.container}>
      <View style={styles.header}>
        <Avatar.Image size={80} source={{ uri: user.avatar }} />
        <Text variant="headlineSmall" style={styles.name}>{user.name}</Text>
        <Text variant="bodyMedium" style={styles.email}>{user.email}</Text>
        <Button mode="contained" style={styles.editButton}>
          Edit Profile
        </Button>
      </View>

      <Divider />

      <List.Section>
        <List.Subheader>Your Relationships</List.Subheader>
        {relationships.map((relationship) => (
          <List.Item
            key={relationship.id}
            title={relationship.name}
            left={() => (
              <Avatar.Image
                size={40}
                source={{ uri: relationship.avatar }}
                style={styles.relationshipAvatar}
              />
            )}
            right={(props) => <List.Icon {...props} icon="chevron-right" />}
          />
        ))}
        <Button
          mode="outlined"
          style={styles.addButton}
          icon="account-plus"
        >
          Add Relationship
        </Button>
      </List.Section>

      <List.Section>
        <List.Subheader>Settings</List.Subheader>
        <List.Item
          title="Notifications"
          left={props => <List.Icon {...props} icon="bell" />}
          right={props => <List.Icon {...props} icon="chevron-right" />}
        />
        <List.Item
          title="Privacy"
          left={props => <List.Icon {...props} icon="shield" />}
          right={props => <List.Icon {...props} icon="chevron-right" />}
        />
        <List.Item
          title="Account"
          left={props => <List.Icon {...props} icon="account" />}
          right={props => <List.Icon {...props} icon="chevron-right" />}
        />
      </List.Section>

      <Button
        mode="outlined"
        style={styles.logoutButton}
        textColor="red"
        onPress={handleLogout}
      >
        Log Out
      </Button>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  header: {
    alignItems: 'center',
    padding: 20,
    backgroundColor: 'white',
  },
  name: {
    marginTop: 10,
  },
  email: {
    marginTop: 4,
    marginBottom: 16,
    color: '#666',
  },
  editButton: {
    marginBottom: 10,
  },
  relationshipAvatar: {
    marginRight: 8,
  },
  addButton: {
    margin: 16,
  },
  logoutButton: {
    margin: 16,
    borderColor: 'red',
  },
}); 