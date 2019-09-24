package parser_test

import (
	"bytes"
	"strings"
	"testing"

	prs "github.com/romshark/llparser/examples/boolexp/parser"
	"github.com/stretchr/testify/require"
)

type Expectation struct {
	AST    string
	Result bool
}

// compareStringifiedAST helps comparing the expected and actual ASTs
// in their stringified form to make display the differences in a human-readable
// way before comparing them recursively
func compareStringifiedAST(
	t *testing.T,
	expected string,
	actual *prs.AST,
) {
	buf := new(bytes.Buffer)
	_, err := actual.Print(prs.ASTPrintOptions{Out: buf})
	require.NoError(t, err)

	// Remove tabs and replace line breaks with whitespaces
	expected = strings.ReplaceAll(expected, "\t", "")
	expected = strings.ReplaceAll(expected, "\n", " ")

	require.Equal(t, expected, buf.String())
}

type Test map[string]Expectation

func (ts Test) Exec(t *testing.T) {
	for expr, expectation := range ts {
		t.Run(expr, func(t *testing.T) {
			pr := prs.NewParser()
			ast, err := pr.Parse("test.boolexp", []rune(expr))
			require.NoError(t, err)
			require.NotNil(t, ast)
			compareStringifiedAST(t, expectation.AST, ast)
			require.Equal(t, expectation.Result, ast.Root.Val())
		})
	}
}

// TestParserPresedence tests the correct order of execution
// of logical operators assuming the following presedence order:
// "!" > "&&" > "||"
func TestParserPresedence(t *testing.T) {
	// If operator presedence is not respected then the following expressions
	// will evaluate to false
	tests := Test{
		"true || !true && false": Expectation{
			Result: true,
			AST: `or{
				const(true)
				and{
					neg{
						const(true)
					}
					const(false)
				}
			}`,
		},
		"true && true || false && false": Expectation{
			Result: true,
			AST: `or{
				and{
					const(true)
					const(true)
				}
				and{
					const(false)
					const(false)
				}
			}`,
		},
	}
	tests.Exec(t)
}

// TestParser tests parsing valid expressions
func TestParser(t *testing.T) {
	tests := Test{
		"true": Expectation{
			Result: true,
			AST:    "const(true)",
		},
		"false": Expectation{
			Result: false,
			AST:    "const(false)",
		},
		"true && true": Expectation{
			Result: true,
			AST: `and{
				const(true)
				const(true)
			}`,
		},
		"true && false": Expectation{
			Result: false,
			AST: `and{
				const(true)
				const(false)
			}`,
		},
		"false && true": Expectation{
			Result: false,
			AST: `and{
				const(false)
				const(true)
			}`,
		},
		"false && false": Expectation{
			Result: false,
			AST: `and{
				const(false)
				const(false)
			}`,
		},
		"true || true": Expectation{
			Result: true,
			AST: `or{
				const(true)
				const(true)
			}`,
		},
		"true || false": Expectation{
			Result: true,
			AST: `or{
				const(true)
				const(false)
			}`,
		},
		"false || true": Expectation{
			Result: true,
			AST: `or{
				const(false)
				const(true)
			}`,
		},
		"false || false": Expectation{
			Result: false,
			AST: `or{
				const(false)
				const(false)
			}`,
		},
		"!false": Expectation{
			Result: true,
			AST: `neg{
				const(false)
			}`,
		},
		"!(false)": Expectation{
			Result: true,
			AST: `neg{
				par{
					const(false)
				}
			}`,
		},
		"(!false)": Expectation{
			Result: true,
			AST: `par{
				neg{
					const(false)
				}
			}`,
		},
		"!(!false)": Expectation{
			Result: false,
			AST: `neg{
				par{
					neg{
						const(false)
					}
				}
			}`,
		},
		"!!false": Expectation{
			Result: false,
			AST: `neg{
				neg{
					const(false)
				}
			}`,
		},
		"!!!false": Expectation{
			Result: true,
			AST: `neg{
				neg{
					neg{
						const(false)
					}
				}
			}`,
		},
		"(true || true && true) || (( true && true ) && false)": Expectation{
			Result: true,
			AST: `or{
				par{
					or{
						const(true)
						and{
							const(true)
							const(true)
						}
					}
				}
				par{
					and{
						par{
							and{
								const(true)
								const(true)
							}
						}
						const(false)
					}
				}
			}`,
		},
	}
	tests.Exec(t)
}
