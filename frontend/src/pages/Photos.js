import React, { Component } from 'react'
import ImageGrid from '../components/ImageGrid'
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
            images: []
        }
    }

    componentDidMount() {
        axios.get(this.props.baseURL+'/photos')
            .then(response => response.data)
            .then(data => this.setState({ images: data.data }))
            .catch(error => console.log(error))
    }

    componentWillUnmount() {

    }

    render() {
        return (
            <div>
                <ImageGrid baseURL={this.props.baseURL} images={this.state.images} />
            </div>
        )
    }
}
