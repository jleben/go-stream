package stream

//

func Const(value Item) Operator {

  work := func (output Writer) {
    status := Ok
    for status == Ok { status = output.Push(value) }
  }

  return Source(work)
}

//

func Series(items... interface{}) Operator {

  work := func (output Writer) {

    // Prepare inputs

    var inputs [] interface {}

    for _, item := range items {
      switch item := item.(type) {
        case Operator:
          inputs = append(inputs, item.Stream())
        default:
          inputs = append(inputs, item)
      }
    }

    // Process inputs one by one

    for _, in := range inputs {

      NextInput:
      switch input := in.(type) {

        case Reader:
          for {
            data, status := input.Pull()
            if status == Closed { break NextInput }
            status = output.Push(data)
            if status == Interrupted { return }
          }

        default:
          status := output.Push(input)
          if status == Interrupted { return }

      }

    }

  }

  return Source(work)
}

//

func Repeat(op Operator, times int) Operator {

  process := func (output Writer, input Reader) Status {
    for {
      item, status := input.Pull()
      if status != Ok { return status }
      status = output.Push(item)
      if status != Ok { return status }
    }
    return Ok
  }

  work := func (output Writer) {
    if times >= 0 {
      for i := 0; i < times; i++ {
        input := op.Stream()
        status := process(output, input)
        if status == Interrupted { return }
      }
    } else {
      for {
        input := op.Stream()
        status := process(output, input)
        if status == Interrupted { return }
      }
    }
  }

  return Source(work)
}

func Join (sources ... Operator) Operator {
  source_done := make(chan struct{})

  forward := func(input Reader, output Writer) {
    for {
      item, status := input.Pull()
      switch status {
        case Ok:
          status = output.Push(item)
          if status == Interrupted { return }
        case Closed:
          source_done <- struct{}{}
          return
      }
    }
  }

  await_sources_done := func() {
    for i := 0; i < len(sources); i++ {
      <- source_done
    }
  }

  work := func(output Writer, inputs ... Reader) {
    for _, input := range inputs {
      go forward(input, output)
    }
    await_sources_done()
  }

  return Filter(work, sources...)
}
