import { LineChart } from "echarts/charts"
import { AxisPointerComponent, BrushComponent, GridComponent, ToolboxComponent, TooltipComponent } from "echarts/components"
import * as echarts from "echarts/core"
import { CanvasRenderer } from "echarts/renderers"

echarts.use([LineChart, GridComponent, AxisPointerComponent, TooltipComponent, BrushComponent, ToolboxComponent, CanvasRenderer])

export default echarts
