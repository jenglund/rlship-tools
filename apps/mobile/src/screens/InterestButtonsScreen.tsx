import React, { useState } from 'react';
import { View, StyleSheet, ScrollView } from 'react-native';
import { Button, Card, FAB, Portal, Modal, TextInput, Text, Switch } from 'react-native-paper';

type InterestButton = {
  id: string;
  name: string;
  defaultDuration: number; // in minutes
  isActive: boolean;
  allowEarlyCancel: boolean;
  isLoud: boolean;
};

export default function InterestButtonsScreen() {
  const [visible, setVisible] = useState(false);
  const [newButtonName, setNewButtonName] = useState('');
  const [newButtonDuration, setNewButtonDuration] = useState('30');
  const [allowEarlyCancel, setAllowEarlyCancel] = useState(true);
  const [isLoud, setIsLoud] = useState(true);

  // Example buttons - in real app, this would come from your backend
  const [buttons] = useState<InterestButton[]>([
    {
      id: '1',
      name: 'Lunch Date',
      defaultDuration: 30,
      isActive: false,
      allowEarlyCancel: true,
      isLoud: true,
    },
    {
      id: '2',
      name: 'Movie Night',
      defaultDuration: 240,
      isActive: true,
      allowEarlyCancel: true,
      isLoud: true,
    },
    {
      id: '3',
      name: 'Quick Coffee',
      defaultDuration: 15,
      isActive: false,
      allowEarlyCancel: false,
      isLoud: false,
    },
  ]);

  const showModal = () => setVisible(true);
  const hideModal = () => setVisible(false);

  const handleCreateButton = () => {
    // Here you would typically make an API call to create the button
    console.log('Creating button:', {
      name: newButtonName,
      defaultDuration: parseInt(newButtonDuration),
      allowEarlyCancel,
      isLoud,
    });
    hideModal();
    setNewButtonName('');
    setNewButtonDuration('30');
    setAllowEarlyCancel(true);
    setIsLoud(true);
  };

  const handlePressButton = (buttonId: string, isActive: boolean) => {
    // Here you would typically make an API call to update the button state
    console.log('Toggling button:', buttonId, !isActive);
  };

  return (
    <View style={styles.container}>
      <ScrollView>
        <View style={styles.grid}>
          {buttons.map((button) => (
            <Card key={button.id} style={styles.card}>
              <Card.Content>
                <Text variant="titleMedium">{button.name}</Text>
                <Text variant="bodySmall">
                  {button.defaultDuration} minutes
                </Text>
                <Text variant="bodySmall">
                  {button.isLoud ? '🔔 Notifications On' : '🔕 Silent Mode'}
                </Text>
              </Card.Content>
              <Card.Actions>
                <Button
                  mode={button.isActive ? 'contained' : 'outlined'}
                  onPress={() => handlePressButton(button.id, button.isActive)}
                >
                  {button.isActive ? 'Active' : 'Inactive'}
                </Button>
              </Card.Actions>
            </Card>
          ))}
        </View>
      </ScrollView>

      <Portal>
        <Modal
          visible={visible}
          onDismiss={hideModal}
          contentContainerStyle={styles.modalContent}
        >
          <Text variant="titleLarge" style={styles.modalTitle}>Create Interest Button</Text>
          <TextInput
            label="Button Name"
            value={newButtonName}
            onChangeText={setNewButtonName}
            style={styles.input}
          />
          <TextInput
            label="Default Duration (minutes)"
            value={newButtonDuration}
            onChangeText={setNewButtonDuration}
            keyboardType="numeric"
            style={styles.input}
          />
          <View style={styles.switchContainer}>
            <Text>Allow Early Cancel</Text>
            <Switch value={allowEarlyCancel} onValueChange={setAllowEarlyCancel} />
          </View>
          <View style={styles.switchContainer}>
            <Text>Send Notifications</Text>
            <Switch value={isLoud} onValueChange={setIsLoud} />
          </View>
          <View style={styles.modalButtons}>
            <Button onPress={hideModal}>Cancel</Button>
            <Button mode="contained" onPress={handleCreateButton}>Create</Button>
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
  grid: {
    padding: 16,
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 16,
  },
  card: {
    width: '47%',
    marginBottom: 0,
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
  switchContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
  },
  modalButtons: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
    gap: 8,
    marginTop: 8,
  },
}); 