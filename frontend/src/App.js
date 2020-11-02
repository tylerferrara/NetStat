import './App.css';

import { BrowserRouter, Route, Switch } from 'react-router-dom';

import Home from "./Components/Login";
import Vote from "./Components/Vote";
import ProtectedRoute from './Components/protectedRoute.js';

function App() {
  return (
      <main>
          <Switch>

            <Route path="/" component={Home} />
            <ProtectedRoute exact={true} path="/" path="/vote" component={Vote} /> 
          </Switch>
      </main>
  )
}

export default App;
