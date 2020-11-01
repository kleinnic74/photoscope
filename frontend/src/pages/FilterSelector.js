import React, { Component } from 'react'
import Timeline from '../components/Timeline'
import PropTypes from 'prop-types'
import axios from 'axios'
import Locations from '../components/Locations'
import Tabs from '../components/Tabs'

export default class FilterSelector extends Component {
    static propTypes = {
        baseURL: PropTypes.string,
        onFilterChanged: PropTypes.func
    }
    static defaultProps = {
        baseURL: window.baseURL,
        onFilterChanged: (url, params) => { }
    }

    constructor(props) {
        super(props)

        this.state = {
            timeline: {},
            locations: {}
        }

        this.onDaySelected = this.onDaySelected.bind(this)
        this.onFilterChange = this.onFilterChange.bind(this)
        this.onPlaceSelected = this.onPlaceSelected.bind(this)

        this.refreshTimeline = this.refreshTimeline.bind(this)
        this.refreshPlaces = this.refreshPlaces.bind(this)
    }

    componentDidMount() {
        this.refreshTimeline()
        this.refreshPlaces()
    }

    refreshTimeline() {
        axios.get(this.props.baseURL + '/timeline/index')
            .then(response => response.data)
            .then(data => {
                console.log("Timeline:", data)
                this.setState({ timeline: data })
            })
            .catch(error => console.log(error))
    }

    refreshPlaces() {
        axios.get(this.props.baseURL + '/geo/index')
            .then(response => response.data)
            .then(data => {
                console.log("Locations:", data)
                this.setState({ locations: data })
            })
            .catch(error => console.log(error))
    }

    onDaySelected(day) {
        console.log("Filter changed to start at day:", day)
        this.props.onFilterChanged('/timeline/photos', { from: day })
    }

    onPlaceSelected(placeID) {
        this.props.onFilterChanged('/geo/photos/byplace/' + placeID)
    }

    onFilterChange(tab) {
        console.log("Filter tab changed", tab)
        switch (tab) {
            case 'Places':
                this.refreshPlaces()
                break
            case 'Timeline':
                this.refreshTimeline()
            default:
                break
        }
    }

    render() {
        return (
            <Tabs onTabChange={this.onFilterChange} >
                <div label="Timeline">
                    <Timeline timeline={this.state.timeline} onDaySelected={this.onDaySelected}></Timeline>
                </div>
                <div label="Places">
                    <Locations locations={this.state.locations} onPlaceSelected={this.onPlaceSelected}></Locations>
                </div>
            </Tabs>
        )
    }
}
