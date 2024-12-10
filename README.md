Pyrene - Fast fake FHIR data fabricator
=======================================

Goals
-----
- Create fake FHIR data much faster than Synthea
- Emphasize simple customization over rigorous realism

Design
------

Like Synthea, Pyrene executes a collection of state machines to simulate various health and healthcare processes
evolving over a patient's lifetime.  These simulated events result in output FHIR resources, which are collected into
bundles for a selected number of randomly-generated patients.

### State machine

Each state machine consists of a variety of states indicating how some process evolves over time:
- `Simple` has no function other than to transition to the next state
- `Guard` does not permit movement to another state until some condition is satisfied
- `Delay` does not move to the next state until a specified amount of simulated time has passed
- `ClinicalEncounter` generates a clinical encounter between the patient and a healthcare entity, along with any outcomes
- `SetState` updates internal state values specific to the state machine
- `End` terminates the state machine execution, and ends the modeled process

States transition from one to the next using a number of different strategies:
- `Direct` simply moves from one state to a defined next state
- `Probability` applies a given probability distribution over the interval \[0, 1\)
- `Condition` tests against a list of conditions and follows the first one to evaluate to true

Note that any time a transition does not provide a valid next state, the `End` state is assumed and the state machine
terminates.

The state machine also tracks the patient's health record as a collection of FHIR resources.  These resources can be
queried within any state machine.

### Configuration

To make the system simple to configure, these state machines should be easily represented in a terse, but human-readable
syntax that is easy to edit.  YAML is a popular and flexible configuration language that can easily represent a complex
state machine.  Each state is named by its key, and must start with a capital letter.  A `Start` state is expected to
exist, and may be of any state type.  The `type` of the state must be specified, and `next` may either be a `string`
reference to another state (a convenient shorthand for the `Direct` transition type) or a full transition object with
specified `type` and other properties as specified by each transition type.  Transitions can be nested, enabling complex
rules to be simply specified without extra states.  States may also include additional optional and/or required
properties depending on each state type.

```yml
Suspected Epilepsy:
  type: Simple
  transition:
    type: Condition
    conditions:
      - 
        condition: Patient.gender = 'male'
        next:
          type: Probability
          split:
            - 
              percent: 55
              next: Seizure Disorder
      - 
        condition: Patient.gender = 'female'
        next:
          type: Probability
          split:
            - 
              percent: 45
              next: Seizure Disorder
```

<details>
<summary>Example: Epilepsy</summary>

```yml
title: Epilepsy
states:
  Start:
    type: Simple
    next:
        type: Probability
        split:
          - 
            percent: 2.2
            next: Ages 0-1
          - 
            percent: 1.95
            next: Ages 1-10
          - 
            percent: 1.95
            next: Ages 10-15
          - 
            pecent: 1.95
            next: Ages 15-55
          - 
            percent: 1.95
            next: Ages 55+
  Ages 0-1:
    type: Delay
    range: 0-1 yr
    next: Suspected Epilepsy
  Ages 1-10:
    type: Delay
    range: 1-10 yr
    next: Suspected Epilepsy
  Ages 10-15:
    type: Delay
    range: 10-15 yr
    next: Suspected Epilepsy
  Ages 15-55:
    type: Delay
    range: 15-55 yr
    next: Suspected Epilepsy
  Ages 55+:
    type: Delay
    range: 55-90 yr
    next: Suspected Epilepsy
  Suspected Epilepsy:
    type: Simple
    transition:
      type: Condition
      conditions:
        - 
          condition: Patient.gender = 'male'
          next:
            type: Probability
            split:
              - 
                percent: 55
                next: Seizure Disorder
        - 
          condition: Patient.gender = 'female'
          next:
            type: Probability
            split:
              - 
                percent: 45
                next: Seizure Disorder
  Seizure Disorder:
    type: ClinicalEncounter
```

</details>

Note that the `End` state does not need to be specified; it is assumed to exist under that name, though it can mostly be
omitted from the configuration.