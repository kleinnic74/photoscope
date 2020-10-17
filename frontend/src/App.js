import React, { Component } from 'react'
import './App.css'
import Photos from './pages/Photos'
import Tasks from './pages/Tasks'
import FilterSelector from './pages/FilterSelector'


export default class App extends Component {

  constructor(props) {
    super(props)

    this.filterChanged = this.filterChanged.bind(this)

    this.state = {
      filter: {path: '/photos', params: {}}
    }
  }

  filterChanged(url, params) {
    console.log("Day selected:", url, params)
    this.setState({filter: {path: url, params: params}})
  }

  render() {
    //  const baseURL = document.baseURI
    const baseURL = "http://localhost:8080"
    console.log("Base URL", baseURL, "filter", this.state.filter)
    return (
      <div className="App">
        <div className="Menu">
          <FilterSelector baseURL={baseURL} onFilterChanged={this.filterChanged} />
        </div>
        <div className="Content">
          <Photos baseURL={baseURL} filter={this.state.filter}/>
        </div>
        <div className="Console">
          <Tasks baseURL={baseURL} />
        </div>
      </div>
    )
  }
}
