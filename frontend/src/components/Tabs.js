import React, { Component } from 'react'
import PropTypes from 'prop-types'
import Tab from './Tab'
import './Tabs.css'

export default class Tabs extends Component {
    static propTypes = {
        children: PropTypes.instanceOf(Array).isRequired,
        onTabChange: PropTypes.func
    }

    constructor(props) {
        super(props)

        this.state = {
            activeTab: this.props.children[0].props.label
        }
    }

    onClickTabItem = (tab) => {
        this.setState({ activeTab: tab })
    }

    componentDidUpdate(prevProps, prevState) {
        if (prevState.activeTab !== this.state.activeTab) {
            this.props.onTabChange(this.state.activeTab)
        }
    }

    render() {
        const {
            onClickTabItem,
            props: {
                children
            },
            state: {
                activeTab
            }
        } = this
        return (
            <div className="Tabs">
                <ol className="Tabs-list">
                    {children.map((child) => {
                        const { label } = child.props
                        return <Tab activeTab={this.state.activeTab}
                            key={label}
                            label={label}
                            onClick={onClickTabItem} />
                    })}
                </ol>
                <div className="Tab-content">
                    {
                        children.map((child) => {
                            if (child.props.label !== activeTab) return undefined
                            return child.props.children
                        })
                    }
                </div>
            </div>
        )
    }
}
