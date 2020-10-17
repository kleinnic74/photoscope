import React, { Component } from 'react'
import PropTypes from 'prop-types'
import './ImageViewer.css'

import Image from './Image'
import classNames from 'classnames'

export class ImageViewer extends Component {
    static propTypes = {
        src: PropTypes.string,
        visible: PropTypes.bool,
        onNext: PropTypes.func,
        onPrev: PropTypes.func
    }

    static defaultProps = {
        onNext: () => { },
        onPrev: () => { }
    }
    constructor(props) {
        super(props)

        this.selfref = React.createRef()

        this.focus = this.focus.bind(this)
        this.handleKeyDown = this.handleKeyDown.bind(this)
    }

    handleKeyDown(ev) {
        switch (ev.keyCode) {
            case 39: //Arrow right
                this.props.onNext()
                break
            case 37: //Arrow left
                this.props.onPrev()
                break
            default:
                break
        }
    }

    componentDidMount() {
        this.focus()
    }

    componentDidUpdate(prevProps, prevState) {
        if (this.props.visible) {
            this.focus()
        }
    }

    focus() {
        this.selfref.current.focus()
    }

    render() {
        var viewerClass = classNames({
            'ImageViewer': true,
            'Visible': this.props.visible,
            'Hidden': !this.props.visible
        })
        return (
            <div ref={this.selfref} tabIndex={0} className={viewerClass} onClick={this.props.onClick} onKeyDown={this.handleKeyDown}>
                <Image src={this.props.src} />
            </div>
        )
    }
}

export default ImageViewer
