import React, { Component } from 'react'
import ImageGrid from '../components/ImageGrid'
import Navbar from '../components/Navbar'

import axios from 'axios'
import PropTypes from 'prop-types'

export default class Photos extends Component {
    static propTypes = {
        baseURL: PropTypes.string
    }

    static defaultProps = {
        baseURL: window.baseURL
    }

    constructor(props) {
        super(props)

        this.state = {
            images: [],
            links: []
        }
        this.onNavClicked = this.onNavClicked.bind(this)
    }

    componentDidMount() {
        axios.get(this.props.baseURL+'/photos')
            .then(response => response.data)
            .then(data => {
                console.log(data.links)
                this.setState({ images: data.data, links: data.links })
            })
            .catch(error => console.log(error))
    }

    componentWillUnmount() {

    }

    onNavClicked(cursor) {
        console.log("Nav clicked")
        axios.get(this.props.baseURL+'/photos', {
            params:{
                c: cursor
            } 
        })
            .then(response => response.data)
            .then(data => {
                console.log(data.links)
                this.setState({ images: data.data, links: data.links })
            })
            .catch(error => console.log(error))
    }

    render() {
        return (
            <div>
                <ImageGrid baseURL={this.props.baseURL} images={this.state.images} />
                <Navbar links={this.state.links} onClick={this.onNavClicked} />
            </div>
        )
    }
}
