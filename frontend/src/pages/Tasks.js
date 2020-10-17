import React, { Component } from 'react'
import PropTypes from 'prop-types'
import TaskList from '../components/TaskList'

import axios from 'axios'

export default class Tasks extends Component {
    static propTypes = {
        baseURL: PropTypes.string
    }

    static defaultProps = {
        baseURL: window.baseURL
    }

    constructor(props) {
        super(props)
        this.state = {
            tasks: [],
            links: []
        }
    } 

    componentDidMount() {
        axios.get(this.props.baseURL+'/tasks')
            .then(response => response.data)
            .then(data => {
                console.log("Tasks:", data)
                this.setState({ tasks: data.data, links: data.links })
            })
            .catch(error => console.log(error))
    }

    render() {
        return (
            <TaskList tasks={this.state.tasks} />
        )
    }
}
