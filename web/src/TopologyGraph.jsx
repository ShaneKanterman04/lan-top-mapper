import React, { useEffect, useMemo, useRef } from 'react'
import cytoscape from 'cytoscape'

function TopologyGraph({ topology, selectedNodeId, onSelectNode, filters, layoutName }) {
  const containerRef = useRef(null)
  const cyRef = useRef(null)

  const elements = useMemo(() => {
    const search = (filters.search || '').toLowerCase().trim()
    const nodes = (topology?.nodes || []).filter((node) => {
      const data = node.data || {}
      if (data.kind === 'gateway') {
        return true
      }
      if (filters.onlineOnly && !data.online) {
        return false
      }
      if (filters.onlyServices && !(data.services || []).length) {
        return false
      }
      if (filters.evidence !== 'all' && !(data.evidence || []).includes(filters.evidence)) {
        return false
      }
      if (filters.deviceType !== 'all' && data.deviceType !== filters.deviceType) {
        return false
      }
      if (filters.confidence !== 'all' && (data.classificationConfidence || 'low') !== filters.confidence) {
        return false
      }
      if (filters.nameSource !== 'all' && (data.nameSource || 'ip') !== filters.nameSource) {
        return false
      }
      if (!search) {
        return true
      }
      return [data.displayName, data.label, data.ip, data.hostname, data.vendor, data.deviceType, data.nameSource]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(search))
    })

    const allowedIDs = new Set(nodes.map((node) => node.data.id))
    const edges = (topology?.edges || []).filter((edge) => {
      const data = edge.data || {}
      return allowedIDs.has(data.source) && allowedIDs.has(data.target)
    })

    return [...nodes, ...edges]
  }, [filters, topology])

  useEffect(() => {
    if (!containerRef.current) {
      return undefined
    }

    const cy = cytoscape({
      container: containerRef.current,
      elements,
      layout: buildLayout(layoutName),
      style: [
        {
          selector: 'node',
          style: {
            'background-color': (ele) => nodeColor(ele.data()),
            label: (ele) => compactLabel(ele.data()),
            color: '#e8eef9',
            'font-size': (ele) => (ele.data().kind === 'gateway' ? 12 : 11),
            'font-weight': (ele) => (ele.data().classificationConfidence === 'high' || ele.data().kind === 'gateway' ? 700 : 600),
            'text-wrap': 'wrap',
            'text-max-width': 92,
            'text-valign': 'bottom',
            'text-margin-y': 8,
            width: (ele) => nodeSize(ele.data()),
            height: (ele) => nodeSize(ele.data()),
            'border-width': (ele) => borderWidth(ele.data()),
            'border-color': (ele) => borderColor(ele.data()),
            'overlay-opacity': 0,
          },
        },
        {
          selector: 'node[kind = "gateway"]',
          style: {
            color: '#06101f',
            'text-margin-y': 10,
          },
        },
        {
          selector: 'node[online = false]',
          style: {
            opacity: 0.55,
          },
        },
        {
          selector: 'edge',
          style: {
            width: 2,
            'line-color': 'rgba(142, 164, 200, 0.45)',
            'target-arrow-shape': 'none',
            'curve-style': 'bezier',
          },
        },
        {
          selector: 'node:selected',
          style: {
            'border-color': '#f8fafc',
            'border-width': 5,
            'shadow-blur': 18,
            'shadow-color': 'rgba(255,255,255,0.35)',
            'shadow-opacity': 1,
          },
        },
      ],
      wheelSensitivity: 0.2,
    })

    cy.on('tap', 'node', (event) => {
      onSelectNode(event.target.data())
    })

    cyRef.current = cy
    window.__LTM_TOPOLOGY_CY__ = cy

    return () => {
      if (window.__LTM_TOPOLOGY_CY__ === cy) {
        delete window.__LTM_TOPOLOGY_CY__
      }
      cy.destroy()
      cyRef.current = null
    }
  }, [elements, layoutName, onSelectNode])

  useEffect(() => {
    const cy = cyRef.current
    if (!cy) {
      return
    }
    cy.layout(buildLayout(layoutName)).run()
    cy.fit(undefined, 32)
  }, [layoutName, elements])

  useEffect(() => {
    const cy = cyRef.current
    if (!cy) {
      return
    }
    if (!selectedNodeId) {
      cy.elements().unselect()
      cy.fit(undefined, 32)
      return
    }
    const node = cy.getElementById(selectedNodeId)
    if (node.length) {
      node.select()
      cy.animate({
        center: { eles: node },
        duration: 250,
      })
    }
  }, [selectedNodeId])

  return <div className="graph-canvas" data-testid="topology-canvas" ref={containerRef} />
}

function buildLayout(layoutName) {
  switch (layoutName) {
    case 'radial':
      return {
        name: 'circle',
        padding: 30,
        spacingFactor: 1.15,
      }
    case 'concentric':
      return {
        name: 'concentric',
        padding: 28,
        spacingFactor: 1.1,
        concentric: (node) => concentricWeight(node.data()),
        levelWidth: () => 1,
      }
    default:
      return {
        name: 'breadthfirst',
        roots: ['gateway'],
        padding: 24,
        spacingFactor: 1.15,
      }
  }
}

function compactLabel(data) {
  const value = data.displayName || data.label || data.hostname || data.ip || 'device'
  return value.length > 18 ? `${value.slice(0, 17)}...` : value
}

function nodeColor(data) {
  if (data.kind === 'gateway') return '#65e2d9'
  switch (data.deviceType) {
    case 'Router or access point':
    case 'Mesh router':
      return '#4ad7a8'
    case 'Computer':
      return '#5d9cff'
    case 'NAS or file server':
      return '#9b7bff'
    case 'Printer':
      return '#f6b84c'
    case 'Media device':
    case 'Streaming device':
      return '#ff8c66'
    case 'Apple device':
      return '#84c5ff'
    case 'Server or appliance':
      return '#6f8bff'
    default:
      return '#64748b'
  }
}

function nodeSize(data) {
  if (data.kind === 'gateway') return 42
  if (data.deviceType === 'Router or access point' || data.deviceType === 'Mesh router') return 32
  if ((data.services || []).length) return 28
  return 24
}

function borderColor(data) {
  switch (data.classificationConfidence) {
    case 'high':
      return '#f8fafc'
    case 'medium':
      return '#cbd5e1'
    default:
      return '#334155'
  }
}

function borderWidth(data) {
  switch (data.classificationConfidence) {
    case 'high':
      return 4
    case 'medium':
      return 3
    default:
      return 2
  }
}

function concentricWeight(data) {
  if (data.kind === 'gateway') return 5
  if (data.deviceType === 'Router or access point' || data.deviceType === 'Mesh router') return 4
  if (data.classificationConfidence === 'high') return 3
  if ((data.services || []).length) return 2
  return 1
}

export default TopologyGraph
