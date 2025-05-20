import React, { useState, useEffect } from 'react';
import { Modal, Button, Form, Alert, Spinner, Card, Row, Col } from 'react-bootstrap';
import { useList } from '../contexts/ListContext';
import { FaRandom, FaDice, FaHistory, FaCheck } from 'react-icons/fa';

const GenerateSelectionModal = ({ show, onHide, listId, listItems = [] }) => {
  const [count, setCount] = useState(1);
  const [excludeRecentlySelected, setExcludeRecentlySelected] = useState(true);
  const [excludePeriod, setExcludePeriod] = useState('30'); // days
  const [selectedItems, setSelectedItems] = useState([]);
  const [selectionHistory, setSelectionHistory] = useState([]);
  const [error, setError] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [hasGenerated, setHasGenerated] = useState(false);
  
  // Load selection history from localStorage when modal opens
  useEffect(() => {
    if (show && listId) {
      const historyKey = `selection_history_${listId}`;
      const savedHistory = localStorage.getItem(historyKey);
      
      if (savedHistory) {
        try {
          setSelectionHistory(JSON.parse(savedHistory));
        } catch (err) {
          console.error('Error parsing selection history:', err);
          setSelectionHistory([]);
        }
      } else {
        setSelectionHistory([]);
      }
      
      // Reset state when modal opens
      setSelectedItems([]);
      setHasGenerated(false);
      setError(null);
    }
  }, [show, listId]);
  
  const handleClose = () => {
    onHide();
  };
  
  const saveToHistory = (items) => {
    const selection = {
      date: new Date().toISOString(),
      items: items.map(item => ({
        id: item.id,
        name: item.name
      }))
    };
    
    const newHistory = [selection, ...selectionHistory].slice(0, 50); // Keep only last 50 selections
    setSelectionHistory(newHistory);
    
    // Save to localStorage
    const historyKey = `selection_history_${listId}`;
    localStorage.setItem(historyKey, JSON.stringify(newHistory));
  };
  
  const getExcludedItemIds = () => {
    if (!excludeRecentlySelected || selectionHistory.length === 0) {
      return [];
    }
    
    const cutoffDate = new Date();
    cutoffDate.setDate(cutoffDate.getDate() - parseInt(excludePeriod, 10));
    
    // Get all item IDs from selections within the exclusion period
    return selectionHistory
      .filter(selection => new Date(selection.date) > cutoffDate)
      .flatMap(selection => selection.items.map(item => item.id));
  };
  
  const generateSelection = () => {
    setIsLoading(true);
    setError(null);
    
    try {
      // Get items to exclude based on history
      const excludedItemIds = getExcludedItemIds();
      
      // Filter available items
      let availableItems = listItems.filter(item => !excludedItemIds.includes(item.id));
      
      if (availableItems.length === 0) {
        // If all items would be excluded, ignore exclusion rules
        availableItems = [...listItems];
        setError('All items were recently selected. Ignoring exclusion rules.');
      }
      
      if (availableItems.length === 0) {
        setError('No items available to select from.');
        setSelectedItems([]);
        setHasGenerated(true);
        return;
      }
      
      // Limit count to available items
      const selectionCount = Math.min(count, availableItems.length);
      
      // Randomly select items
      const selected = [];
      const itemsCopy = [...availableItems];
      
      for (let i = 0; i < selectionCount; i++) {
        const randomIndex = Math.floor(Math.random() * itemsCopy.length);
        selected.push(itemsCopy[randomIndex]);
        itemsCopy.splice(randomIndex, 1); // Remove selected item to avoid duplicates
      }
      
      setSelectedItems(selected);
      saveToHistory(selected);
      setHasGenerated(true);
    } catch (err) {
      console.error('Error generating selection:', err);
      setError('Failed to generate selection. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };
  
  const handleCountChange = (e) => {
    const value = parseInt(e.target.value, 10);
    setCount(isNaN(value) || value < 1 ? 1 : value);
  };
  
  const formatDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
  };
  
  const renderSelectionHistory = () => {
    if (selectionHistory.length === 0) {
      return <p className="text-muted">No selection history available.</p>;
    }
    
    return (
      <div className="selection-history">
        <h6 className="mb-3">Recent Selections</h6>
        {selectionHistory.slice(0, 5).map((selection, index) => (
          <div key={index} className="history-item mb-2">
            <small className="text-muted d-block">{formatDate(selection.date)}</small>
            <p className="mb-0">
              {selection.items.map(item => item.name).join(', ')}
            </p>
            <hr className="my-2" />
          </div>
        ))}
      </div>
    );
  };
  
  return (
    <Modal show={show} onHide={handleClose} centered size="lg">
      <Modal.Header closeButton>
        <Modal.Title><FaRandom className="me-2" />Generate Random Selection</Modal.Title>
      </Modal.Header>
      
      <Modal.Body>
        {error && (
          <Alert variant="warning">{error}</Alert>
        )}
        
        <Row>
          <Col md={6}>
            <Form>
              <Form.Group className="mb-3">
                <Form.Label>Number of items to select</Form.Label>
                <Form.Control
                  type="number"
                  min="1"
                  value={count}
                  onChange={handleCountChange}
                  disabled={isLoading}
                />
                <Form.Text className="text-muted">
                  Choose how many items you want to select
                </Form.Text>
              </Form.Group>
              
              <Form.Group className="mb-3">
                <Form.Check
                  type="checkbox"
                  label="Exclude recently selected items"
                  checked={excludeRecentlySelected}
                  onChange={(e) => setExcludeRecentlySelected(e.target.checked)}
                  disabled={isLoading}
                />
                
                {excludeRecentlySelected && (
                  <Form.Group className="mt-2">
                    <Form.Label>Exclude items selected in the past:</Form.Label>
                    <Form.Select
                      value={excludePeriod}
                      onChange={(e) => setExcludePeriod(e.target.value)}
                      disabled={isLoading}
                    >
                      <option value="7">7 days</option>
                      <option value="14">14 days</option>
                      <option value="30">30 days</option>
                      <option value="60">60 days</option>
                      <option value="90">90 days</option>
                    </Form.Select>
                  </Form.Group>
                )}
              </Form.Group>
              
              <Button 
                variant="primary" 
                onClick={generateSelection} 
                disabled={isLoading}
                className="w-100"
              >
                {isLoading ? (
                  <>
                    <Spinner as="span" animation="border" size="sm" role="status" aria-hidden="true" className="me-2" />
                    Generating...
                  </>
                ) : (
                  <>
                    <FaDice className="me-2" /> Generate Selection
                  </>
                )}
              </Button>
            </Form>
            
            <div className="mt-4">
              {renderSelectionHistory()}
            </div>
          </Col>
          
          <Col md={6}>
            <h5 className="mb-3">Selection Result</h5>
            
            {hasGenerated ? (
              selectedItems.length > 0 ? (
                <div className="selection-results">
                  {selectedItems.map(item => (
                    <Card key={item.id} className="mb-2 shadow-sm">
                      <Card.Body>
                        <Card.Title className="h6">{item.name}</Card.Title>
                        {item.description && (
                          <Card.Text className="small text-muted">
                            {item.description}
                          </Card.Text>
                        )}
                      </Card.Body>
                    </Card>
                  ))}
                </div>
              ) : (
                <Alert variant="info">
                  No items were selected. Try adjusting your criteria.
                </Alert>
              )
            ) : (
              <div className="text-center py-5 text-muted">
                <FaDice size={48} className="mb-3 text-secondary" />
                <p>Click "Generate Selection" to randomly select items from your list.</p>
              </div>
            )}
          </Col>
        </Row>
      </Modal.Body>
      
      <Modal.Footer>
        <Button variant="secondary" onClick={handleClose}>
          Close
        </Button>
      </Modal.Footer>
    </Modal>
  );
};

export default GenerateSelectionModal; 