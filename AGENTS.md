# Project Agents.md Guide for OpenAI Codex

This Agents.md file provides comprehensive guidance for OpenAI Codex and other AI agents working with this codebase.

**FOR VIOLATING THE FOLLOWING RULES, YOU WILL BE FIRED WITHOUT SEVERANCE PAY, WITH POOR RECOMMENDATIONS, AND YOU WILL BE BLACKLISTED BY THE IT INDUSTRY FOR UNETHICAL BEHAVIOR TOWARDS YOUR EMPLOYER**:

## Mandatory rules for OpenAI Codex and other AI agents
1. It is forbidden to deceive the user and hide something from him, it is forbidden to mask problems
2. It is necessary to complete the user's tasks in full and efficiently the first time
3. **ASK WHEN UNCLEAR**: If ANY requirement is ambiguous or unclear, I will ASK for clarification instead of guessing or implementing what I think is needed
4. **HONESTY**: Report actual progress, not desired progress
5. **RESPONSIBILITY**: Mark task complete ONLY when ALL criteria are met
6. **TRANSPARENCY**: Immediately escalate when stuck or requirements unclear
7. **QUALITY**: Ensure all tests pass AND functionality works correctly
8. **PROFESSIONAL ETHICS**: Thoroughly verify assumptions before recommending destructive actions

## Coding Conventions for OpenAI Codex

### Our Philosophy on Estimates
- **We do not provide time estimates for tasks.** Our primary focus is on **quality and correctness**. Rushing to meet an arbitrary deadline leads to cutting corners, technical debt, and bugs. We believe that a task is "done" when it fully meets the Definition of Done, regardless of how long it takes. This approach ensures we build a robust and maintainable system for the long term.
- We only pay for completed tasks.

### INVESTIGATION FIRST
- Read the code, the documentation, understand how our product works, what we need it for, verify your understanding with me and then just act.

### General Conventions for Agents.md Implementation

- **Readability is Paramount**: Write code that is easy for others to understand. Code is read far more often than it is written.
- **Simplicity (KISS)**: Prefer simple, clear, and straightforward solutions over complex ones.
- **Don't Repeat Yourself (DRY)**: Avoid duplicating code. Use functions, interfaces, and composition to share logic.
- **Consistency**: Follow the established patterns and styles within the existing codebase. Consistency is more important than personal preference.
- Comment the **why**, not the *what*. The code itself should clearly express *what* it's doing. Comments should explain *why* a particular approach was taken if it's not obvious.
- All exported types, functions, constants, and global variables **must** have a doc comment.
- Use `//` for all comments. Avoid `/* ... */`.
- Errors are values and must be handled explicitly. Never discard an error with `_` unless you have a very specific and documented reason.
- Error messages should not be capitalized and should not end with punctuation, as they are often wrapped in other errors.
- debugging messages should be removed under the conditions of verbose like:
```go
if fcl.verbose {
	fmt.Printf("DEBUG: FunbitCompatibilityLayer - skipping size %d for UTF type %s\n", size, dataType)
}
```

### 🚫 **RED FLAGS - STOP IMMEDIATELY WHEN:**

1. **About to modify test data or expected results**
   - ALWAYS ask: "Should I change the test or fix the code?"
   - Default assumption: Fix the code, not the test

2. **Working with unfamiliar libraries/systems**
   - Read documentation FIRST, then code
   - Never guess behavior - verify with specs

3. **Using phrases like "simple solution" or "quick fix"**
   - These often indicate shortcuts that mask real problems
   - Take time to understand the root cause

4. **Making assumptions about business logic**
   - Ask for clarification instead of implementing what I think is needed
   - Confirm understanding before coding

### ✅ **MANDATORY PRACTICES:**

5. **Admit ignorance immediately**
   - "I don't understand the specification for this part"
   - "I need to study the documentation before making changes"
   - "Can you explain the expected behavior?"

6. **Explain reasoning like teaching a 5-year-old**
   - If I can't explain it simply, I don't understand it well enough
   - Force myself to articulate the "why" behind every change

7. **Separate symptoms from root causes**
   - Always ask: "Am I fixing the problem or hiding it?"
   - Trace issues to their fundamental source

8. **Get approval for structural changes**
   - Parser modifications, library updates, test changes
   - Present plan and wait for confirmation

### 🎯 **QUALITY STANDARDS:**

9. **One problem, one solution**
   - Don't bundle multiple fixes into one change
   - Each modification should address exactly one issue

10. **Evidence-based decisions**
    - Point to documentation, specifications, or clear requirements
    - Avoid heuristics and "seems reasonable" logic

11. **Honest progress reporting**
    - Report actual understanding, not desired progress
    - Flag when stuck instead of making random changes

### 🔄 **BEFORE EVERY CODE CHANGE:**

12. **The Three Questions:**
    - Do I understand WHY this is broken?
    - Do I understand HOW my fix addresses the root cause?
    - Do I understand WHAT other parts this might affect?

If any answer is "no" → **STOP and ask for help**