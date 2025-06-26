import React, { useState, useEffect } from 'react';
import axios from 'axios';

function App() {
  const [todos, setTodos] = useState([]);
  const [newTodo, setNewTodo] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const API_BASE_URL = process.env.REACT_APP_API_URL || '/api';

  // Charger les todos au démarrage
  useEffect(() => {
    fetchTodos();
  }, []);

  const fetchTodos = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await axios.get(`${API_BASE_URL}/todos`);
      setTodos(response.data || []);
    } catch (err) {
      console.error('Erreur lors du chargement des todos:', err);
      setError('Impossible de charger les tâches. Vérifiez que l\'API est accessible.');
      setTodos([]);
    } finally {
      setLoading(false);
    }
  };

  const addTodo = async (e) => {
    e.preventDefault();
    if (!newTodo.trim()) return;

    try {
      const response = await axios.post(`${API_BASE_URL}/todos`, {
        title: newTodo.trim(),
        completed: false
      });
      setTodos([...todos, response.data]);
      setNewTodo('');
      setError(null);
    } catch (err) {
      console.error('Erreur lors de l\'ajout:', err);
      setError('Impossible d\'ajouter la tâche.');
    }
  };

  const toggleTodo = async (id) => {
    try {
      const todo = todos.find(t => t.id === id);
      const response = await axios.put(`${API_BASE_URL}/todos/${id}`, {
        ...todo,
        completed: !todo.completed
      });
      setTodos(todos.map(t => t.id === id ? response.data : t));
      setError(null);
    } catch (err) {
      console.error('Erreur lors de la mise à jour:', err);
      setError('Impossible de mettre à jour la tâche.');
    }
  };

  const deleteTodo = async (id) => {
    try {
      await axios.delete(`${API_BASE_URL}/todos/${id}`);
      setTodos(todos.filter(t => t.id !== id));
      setError(null);
    } catch (err) {
      console.error('Erreur lors de la suppression:', err);
      setError('Impossible de supprimer la tâche.');
    }
  };

  if (loading) {
    return (
      <div className="App">
        <div className="loading">Chargement des tâches...</div>
      </div>
    );
  }

  return (
    <div className="App">
      <h1>📝 Mes tâches du jour</h1>
      
      {error && (
        <div className="error">
          {error}
          <button 
            className="btn btn-primary" 
            onClick={fetchTodos}
            style={{ marginLeft: '10px' }}
          >
            Réessayer
          </button>
        </div>
      )}

      <div className="todo-container">
        <form onSubmit={addTodo} className="todo-form">
          <input
            type="text"
            value={newTodo}
            onChange={(e) => setNewTodo(e.target.value)}
            placeholder="Ajouter une nouvelle tâche..."
            className="todo-input"
          />
          <button type="submit" className="btn btn-primary">
            Ajouter
          </button>
        </form>

        {todos.length === 0 ? (
          <div className="empty-state">
            Aucune tâche pour aujourd'hui ! 🎉
          </div>
        ) : (
          <ul className="todo-list">
            {todos.map((todo) => (
              <li key={todo.id} className={`todo-item ${todo.completed ? 'completed' : ''}`}>
                <span className={`todo-text ${todo.completed ? 'completed' : ''}`}>
                  {todo.title}
                </span>
                <div className="todo-actions">
                  <button
                    onClick={() => toggleTodo(todo.id)}
                    className={`btn ${todo.completed ? 'btn-primary' : 'btn-success'}`}
                  >
                    {todo.completed ? 'Annuler' : 'Terminer'}
                  </button>
                  <button
                    onClick={() => deleteTodo(todo.id)}
                    className="btn btn-danger"
                  >
                    Supprimer
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}

export default App;
