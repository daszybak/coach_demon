import { useEffect, useState } from 'react'
import './App.css'

interface Statement {
    problemID: string
    statement: string
}

interface Summary {
    problemID: string
    feedback: string
    proof: string
    optimalMetaCognition: string
}

function App() {
    const [statements, setStatements] = useState<Statement[]>([])
    const [summaries, setSummaries] = useState<Record<string, Summary>>({})
    const [expandedProblemID, setExpandedProblemID] = useState<string | null>(null)
    const [loading, setLoading] = useState(false)

    useEffect(() => {
        const fetchStatements = async () => {
            try {
                const res = await fetch('http://localhost:12345/statements')
                const data = await res.json()
                setStatements(data || [])
            } catch (err) {
                console.error('Failed to load statements', err)
            }
        }
        fetchStatements()
    }, [])

    const toggleSummary = async (problemID: string) => {
        const isExpanded = expandedProblemID === problemID
        setExpandedProblemID(isExpanded ? null : problemID)

        if (!isExpanded && !summaries[problemID]) {
            setLoading(true)
            try {
                const res = await fetch(`http://localhost:12345/summary/${problemID}`)
                const data = await res.json()
                setSummaries((prev) => ({ ...prev, [problemID]: data }))
            } catch (err) {
                console.error('Failed to load summary', err)
            } finally {
                setLoading(false)
            }
        }
    }

    return (
        <div className="app">
            <div className="card">
                <h1 className="card-title">Problem Statements</h1>

                <ul className="statement-list">
                    {statements.map(({ problemID, statement }) => (
                        <li key={problemID} className="statement-item">
                            <button
                                className={`statement-button ${expandedProblemID === problemID ? 'active' : ''}`}
                                onClick={() => toggleSummary(problemID)}
                            >
                                {statement}
                            </button>

                            {expandedProblemID === problemID && (
                                <div className="summary-panel">
                                    {loading ? (
                                        <div className="loader"></div>
                                    ) : (
                                        summaries[problemID] && (
                                            <>
                                                <p><strong>Feedback:</strong> {summaries[problemID].feedback}</p>
                                                <p><strong>Proof:</strong> {summaries[problemID].proof}</p>
                                                <p><strong>Optimal MetaCognition:</strong> {summaries[problemID].optimalMetaCognition}</p>
                                            </>
                                        )
                                    )}
                                </div>
                            )}
                        </li>
                    ))}
                </ul>
            </div>
        </div>
    )
}

export default App
