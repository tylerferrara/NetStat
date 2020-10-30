import './App.css';

import { BrowserRouter, Route, Switch } from 'react-router-dom';

import Home from "./Components/Login";
import Vote from "./Components/Vote";

function App() {
  return (
      <main>
          <Switch>

            <Route path="/" component={Home} exact />
            <Route path="/vote" component={Vote} />
              
          </Switch>
      </main>
  )
}

export default App;
