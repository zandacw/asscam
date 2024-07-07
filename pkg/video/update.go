package video

func (newFrame Frame) diff(oldFrame Frame) updates {

	if oldFrame == nil {
		return nil
	}

	nfCols := len(newFrame)
	var nfRows int
	if nfCols > 0 {
		nfRows = len(newFrame[0])
	}
	ofCols := len(oldFrame)
	var ofRows int
	if ofCols > 0 {
		ofRows = len(oldFrame[0])
	}

	if nfRows != ofRows || nfCols != ofCols {
		return nil
	}

	var updates []update
	for rowIdx := range newFrame {
		for colIdx := range newFrame[rowIdx] {
			oldChar := oldFrame[rowIdx][colIdx]
			newChar := newFrame[rowIdx][colIdx]
			if newChar != oldChar {
				updates = append(
					updates,
					update{rowIdx: rowIdx, colIdx: colIdx, char: newChar},
				)
			}
		}
	}
	return updates
}

func (ups updates) do() {
	for _, update := range ups {
		moveAndWrite(update.rowIdx, update.colIdx, update.char)
	}
}
