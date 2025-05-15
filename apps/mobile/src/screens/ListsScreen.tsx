import React, { useState } from 'react';
import { View, StyleSheet, ScrollView } from 'react-native';
import { List, FAB, Portal, Modal, TextInput, Button, Text } from 'react-native-paper';

type ActivityList = {
  id: string;
  name: string;
  type: string;
  itemCount: number;
};

export default function ListsScreen() {
  const [visible, setVisible] = useState(false);
  const [newListName, setNewListName] = useState('');
  const [newListType, setNewListType] = useState('');

  // Example lists - in real app, this would come from your backend
  const [lists] = useState<ActivityList[]>([
    { id: '1', name: 'Favorite Restaurants', type: 'restaurants', itemCount: 12 },
    { id: '2', name: 'Hiking Trails', type: 'outdoor', itemCount: 5 },
    { id: '3', name: 'Movie Watchlist', type: 'entertainment', itemCount: 8 },
    { id: '4', name: 'Date Night Ideas', type: 'activities', itemCount: 15 },
  ]);

  const showModal = () => setVisible(true);
  const hideModal = () => setVisible(false);

  const handleCreateList = () => {
    // Here you would typically make an API call to create the list
    console.log('Creating list:', { name: newListName, type: newListType });
    hideModal();
    setNewListName('');
    setNewListType('');
  };

  return (
    <View style={styles.container}>
      <ScrollView>
        <List.Section>
          <List.Subheader>Your Activity Lists</List.Subheader>
          {lists.map((list) => (
            <List.Item
              key={list.id}
              title={list.name}
              description={`${list.itemCount} items`}
              left={props => <List.Icon {...props} icon="format-list-bulleted" />}
              right={props => <List.Icon {...props} icon="chevron-right" />}
              onPress={() => console.log('Navigate to list:', list.id)}
            />
          ))}
        </List.Section>
      </ScrollView>

      <Portal>
        <Modal
          visible={visible}
          onDismiss={hideModal}
          contentContainerStyle={styles.modalContent}
        >
          <Text variant="titleLarge" style={styles.modalTitle}>Create New List</Text>
          <TextInput
            label="List Name"
            value={newListName}
            onChangeText={setNewListName}
            style={styles.input}
          />
          <TextInput
            label="List Type"
            value={newListType}
            onChangeText={setNewListType}
            style={styles.input}
          />
          <View style={styles.modalButtons}>
            <Button onPress={hideModal}>Cancel</Button>
            <Button mode="contained" onPress={handleCreateList}>Create</Button>
          </View>
        </Modal>
      </Portal>

      <FAB
        icon="plus"
        style={styles.fab}
        onPress={showModal}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  fab: {
    position: 'absolute',
    margin: 16,
    right: 0,
    bottom: 0,
  },
  modalContent: {
    backgroundColor: 'white',
    padding: 20,
    margin: 20,
    borderRadius: 8,
  },
  modalTitle: {
    marginBottom: 20,
  },
  input: {
    marginBottom: 16,
  },
  modalButtons: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
    gap: 8,
  },
}); 