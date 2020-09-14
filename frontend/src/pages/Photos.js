import React, { Component } from 'react'
import ImageGrid from '../components/ImageGrid'
import Navbar from '../components/Navbar'
import ImageViewer from '../components/ImageViewer'

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
        this.fetchImages()
    }

    onNavClicked(cursor) {
        this.fetchImages(cursor)
    }

    absURL(part) {
        return `${this.props.baseURL}${part}`
    }

    fetchImages(cursor, showFirst) {
        const params = cursor ? { c: cursor } : {}
        axios.get(this.props.baseURL + '/photos', {
            params: params
        })
            .then(response => response.data)
            .then(data => {
                this.setState({
                    images: data.data,
                    links: data.links?.reduce((map,l) => {
                        map[l.name] = l
                        return map
                    }, {}),
                    showImage: showFirst,
                    index: 0,
                    image: this.absURL(data.data[0].links.view)
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
                this.fetchImages(prevState.links.next.href,true)
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
            var prev = prevState.index-1
            if (prev < 0) {
                if (prevState.links.previous) {
                    this.fetchImages(prevState.links.previous.href, true)
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
            <div className="Content">
                <Navbar links={this.state.links} onClick={this.onNavClicked} />
                <ImageGrid baseURL={this.props.baseURL} images={this.state.images} onShow={this.showImage} />
                <Navbar links={this.state.links} onClick={this.onNavClicked} />
                <ImageViewer src={this.state.image} 
                    visible={this.state.showImage} 
                    onClick={this.toggleImage} 
                    onNext={this.showNext} 
                    onPrev={this.showPrevious}/>
            </div>
        )
    }
}
