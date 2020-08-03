import React, { Component } from 'react'
import PropTypes from 'prop-types'
import Image from './Image'
import './ImageGrid.css'

export default class ImageGrid extends Component {
    static propTypes = {
        baseURL: PropTypes.string.isRequired,
        rows: PropTypes.number.isRequired,
        cols: PropTypes.number.isRequired,
        images: PropTypes.array,
    }

    static defaultProps = {
        rows: 5,
        cols: 5,
        images: []
    }

    render() {
        const rows = Array(this.props.rows).fill().map((_, i) =>
            <tr key={i}>
                {Array(this.props.cols).fill().map((_, j) => {
                    const index = i*this.props.cols+j
                    if (index < this.props.images.length) {
                        const img = this.props.images[index]
                        const src = `${this.props.baseURL}${img.links.thumb}`
                        return <td key={j}><Image src={src} alt={img.name} /></td>
                    } else {
                        return <td key={j}></td>
                    }
                })}
            </tr>
        )
        return (
            <table className="ImageGrid">
                <tbody>
                    {rows}
                </tbody>
            </table>
        )
    }
}
