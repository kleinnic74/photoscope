import React, { Component } from 'react'
import PropTypes from 'prop-types'
import './Tree.css'

export default class Locations extends Component {
    static propTypes = {
        locations: PropTypes.object,
        onPlaceSelected: PropTypes.func
    }

    static defaultProps = {
        locations: { countries: [] },
        onPlaceSelected: (p) => {}
    }

    clickHandler = (placeId) => {
        this.props.onPlaceSelected(placeId)
    }

    render() {
        const {
            props: {
                locations: {
                    countries
                }
            }
        } = this


        return (
            <div className="Tree">
                <ul className="Tree-node">
                    {countries.map((c) => <Country key={c.ciso} country={c} onClick={this.clickHandler} />)}
                </ul>
            </div>
        )
    }
}

class Country extends Component {

    clickHandler = (placeId) => {
        this.props.onClick(placeId)
    }

    render() {
        const {
            props: {
                country: {
                    ciso,
                    country,
                    places }
            }
        } = this
        const leaves = places.map((p,i) => {
            return (<Leaf key={i} label={p.city || p.zip} id={ciso+"-"+p.zip} onClick={this.clickHandler}/>)
        })
        return (
            <li>{country}
                <ul>
                    {leaves}
                </ul>
            </li>
        )
    }
}

class Leaf extends Component {
    static propTypes = {
        id: PropTypes.string.isRequired,
        label: PropTypes.string.isRequired,
        onClicked: PropTypes.func
    }

    clickHandler = (ev) => {
        ev.stopPropagation()
        this.props.onClick(this.props.id)
    }

    render() {
        return (<li onClick={this.clickHandler}>{this.props.label}</li>)
    }
}
