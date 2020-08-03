import React from 'react';
import './App.css';
import {
  BrowserRouter as Router,
  Switch,
  Link,
  Route,
  Redirect
} from 'react-router-dom';
import Photos from './pages/Photos';

function App() {
  return (
    <Router>
    <div className="App">
      <div className="Menu">
        <nav>
          <ul>
            <li><Link to="/photos">Photos</Link></li>
            <li><Link to="/tasks">Tasks</Link></li>
          </ul>
        </nav>
      </div>
      <div className="Content">
        <Switch>
          <Route path="/photos">
            <Photos baseURL="http://localhost:8080" />
          </Route>
          <Route path="/tasks">
            <div>Current tasks...</div>
          </Route>
          <Route path="/">
            <Redirect to="/photos"/>
          </Route>
        </Switch>
      </div>
    </div>
    </Router>
  );
}

export default App;
