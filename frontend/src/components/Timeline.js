import React, { Component } from 'react'
import PropTypes from 'prop-types'
import './Tree.css'

export default class Timeline extends Component {
    static propTypes = {
        timeline: PropTypes.object,
        onDaySelected: PropTypes.func
    }

    render() {
        const view = this.props.timeline.years?.map((y) => <Year key={y.year} year={y} onDaySelected={this.props.onDaySelected}/>)
        return (
            <div className="Tree">
                <ul className="Tree-node">
                    {view}
                </ul>
            </div>
        )
    }
}


function Year(props) {
    const months = props.year.months?.map((m) => <Month key={m.month} month={m} onDaySelected={props.onDaySelected}/>)
    return (
        <li>
            {props.year.year}
            <ul className="Tree-node">{months}</ul>
        </li>
    )
}

function Month(props) {
    const days = props.month.days?.map((d) => <li key={d.key} onClick={ev => props.onDaySelected(d.key)}>{d.key}</li>)
    return (<li>{props.month.month}
        <ul className="Tree-leaf">{days}</ul>
    </li>)
}