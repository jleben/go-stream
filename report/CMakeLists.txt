
add_custom_target( report
  COMMAND pdflatex
    -interaction nonstopmode
    "${CMAKE_CURRENT_SOURCE_DIR}/report.tex"
)
