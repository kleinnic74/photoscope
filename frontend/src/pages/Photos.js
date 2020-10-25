import React, { Component } from 'react'
import ImageGrid from '../components/ImageGrid'
import Navbar from '../components/Navbar'
import ImageViewer from '../components/ImageViewer'

import axios from 'axios'
import PropTypes from 'prop-types'

import './Photos.css'

export default class Photos extends Component {
    static propTypes = {
        baseURL: PropTypes.string,
        filter: PropTypes.object
    }

    static defaultProps = {
        baseURL: window.baseURL,
        filter: { path: '/photos', params: {} }
    }

    constructor(props) {
        super(props)

        this.state = {
            images: [],
            links: {},
            showImage: false,
            index: 0
        }
        this.onNavClicked = this.onNavClicked.bind(this)
        this.showImage = this.showImage.bind(this)
        this.toggleImage = this.toggleImage.bind(this)
        this.showNext = this.showNext.bind(this)
        this.showPrevious = this.showPrevious.bind(this)
        this.absURL = this.absURL.bind(this)
    }

    componentDidMount() {
        const size = {
            width: this.toplevel.parentElement.clientWidth,
            height: this.toplevel.parentElement.clientHeight
        }
        const maxImages = { x: Math.floor((size.width - 20) / (128 + 10)), y: Math.floor((size.height - 20) / (128 + 10)) }
        console.log("Photo page size:", size, "Max images:", maxImages)
        this.setState({ maxImages: maxImages.x * maxImages.y })
        this.fetchImages()
    }

    componentDidUpdate(prevProps, prevState) {
        if (this.props.filter !== prevProps.filter ||
            this.props.params !== prevProps.params ||
            this.state.maxImages !== prevState.maxImages) {
            console.log("Grid filter changed: ", prevProps.filter.path, "=>", this.props.filter.path)
            this.fetchImages()
        }
    }

    onNavClicked(cursor) {
        this.fetchImages(cursor)
    }

    absURL(part) {
        return `${this.props.baseURL}${part}`
    }

    fetchImages(cursor, showImage, index) {
        let params = { ...this.props.filter.params }
        if (cursor) {
            params = { c: cursor, ...params}
        }
        if (this.state.maxImages) {
            params = { p: this.state.maxImages, ...params }
        }
        console.log("Fetching images, params:", params)
        axios.get(this.props.baseURL + this.props.filter.path, {
            params: params
        })
            .then(response => response.data)
            .then(data => {
                const i = index < 0 ? data.data.length - 1 : index || 0
                this.setState({
                    images: data.data,
                    links: data.links?.reduce((map, l) => {
                        map[l.name] = l
                        return map
                    }, {}),
                    showImage: showImage,
                    index: i,
                    image: this.absURL(data.data[i].links.view)
                })
            })
            .catch(error => console.log(error))
    }

    showImage(img, index) {
        this.setState({ image: img, showImage: true, index: index })
    }

    showNext() {
        this.setState((prevState) => {
            var next = prevState.index + 1
            if (next >= prevState.images.length) {
                this.fetchImages(prevState.links.next.href, true, 0)
                return {}
            }
            var img = this.absURL(prevState.images[next].links.view)
            return {
                image: img,
                showImage: true,
                index: next,
            }
        })
    }

    showPrevious() {
        this.setState((prevState) => {
            var prev = prevState.index - 1
            if (prev < 0) {
                if (prevState.links.previous) {
                    this.fetchImages(prevState.links.previous.href, true, -1)
                }
                return {}
            }
            var img = this.absURL(prevState.images[prev].links.view)
            return {
                image: img,
                showImage: true,
                index: prev
            }
        })
    }

    toggleImage() {
        this.setState((state, props) => ({
            showImage: !state.showImage,
        }))
    }

    render() {
        return (
            <div id="PhotoPage" ref={(element) => { this.toplevel = element }}>
                <Navbar id="Navtop" links={this.state.links} onClick={this.onNavClicked} />
                <ImageGrid id="Imagegrid" baseURL={this.props.baseURL} images={this.state.images} onShow={this.showImage} />
                <Navbar id="Navbottom" links={this.state.links} onClick={this.onNavClicked} />
                <ImageViewer src={this.state.image}
                    visible={this.state.showImage}
                    onClick={this.toggleImage}
                    onNext={this.showNext}
                    onPrev={this.showPrevious} />
            </div>
        )
    }
}
