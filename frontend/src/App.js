import React from 'react'
import './App.css'
import {
  BrowserRouter as Router,
  Switch,
  Link,
  Route,
  Redirect
} from 'react-router-dom'
import Photos from './pages/Photos'
import Tasks from './pages/Tasks'

function App() {
//  const baseURL = document.baseURI
  const baseURL = "http://localhost:8080"
  console.log("Base URL", baseURL)
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
        <Switch>
          <Route path="/photos">
            <Photos baseURL={baseURL} />
          </Route>
          <Route path="/tasks">
            <Tasks baseURL={baseURL} />
          </Route>
          <Route path="/">
            <Redirect to="/photos"/>
          </Route>
        </Switch>
    </div>
    </Router>
  );
}

export default App;
