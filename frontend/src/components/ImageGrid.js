import React, { Component } from 'react'
import PropTypes from 'prop-types'
import Image from './Image'
import './ImageGrid.css'

export default class ImageGrid extends Component {
    static propTypes = {
        baseURL: PropTypes.string.isRequired,
        images: PropTypes.array,
    }

    static defaultProps = {
        rows: 5,
        cols: 5,
        images: []
    }

    render() {
        const images = this.props.images.map((i, idx) => {
            const thumb = `${this.props.baseURL}${i.links.thumb}`
            const img = `${this.props.baseURL}${i.links.view}`
            return <Image key={thumb} src={thumb} alt={i.name} onClick={(ev) => this.props.onShow(img, idx)}/>
        })
        return (
            <div className="ImageGrid">
                    {images}
            </div>
        )
    }
}
