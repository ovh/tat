---
title: "Dashing View"
weight: 6
toc: true
prev: "/tatwebui/pastatview"
next: "/tatwebui/releaseview"

---


## Usage

This screenshot

![Dashing View](/imgs/tatwebui-dashing-view.png?width=80%)

can be done by sending info with tatcli :

```bash
tatcli msg add /Internal/YourTopic/Dashing "#monitoring" \
  --label="#eeeee;border-width:0px" \
  --label="#eeeee;height:110px" \
  --label="#eeeee;hide-bottom" \
  --label="#eeeee;order:1" \
  --label="#eeeee;value:MyDashing" \
  --label="#eeeee;width:96%" \
  --label="#ffffff;color" \
  --label="#161d23;bg-color"

tatcli msg add /Internal/YourTopic/Dashing "#monitoring" \
  --label="#93c47d;bg-color" \
  --label="#eeeee;border-width:0px" \
  --label="#eeeee;height:5px" \
  --label="#eeeee;hide-bottom" \
  --label="#eeeee;order:2" \
  --label="#eeeee;value" \
  --label="#eeeee;value" \
  --label="#eeeee;value" \
  --label="#eeeee;color" \
  --label="#eeeee;width:96.3%"

tatcli msg add /Internal/YourTopic/Dashing "#monitoring" \
  --label="#161d23;bg-color" \
  --label="#eeeee;border-width:0px" \
  --label="#eeeee;color" \
  --label="#eeeee;height:20px" \
  --label="#eeeee;hide-bottom" \
  --label="#eeeee;order:3" \
  --label="#eeeee;width:96%"

tatcli msg add /Internal/YourTopic/Dashing "#monitoring #YourTeam #item:OCO_Critical" \
  --label="#e8f1f4;bg-color" \
  --label="#ce352c;color" \
  --label="#eeeee;order:11" \
  --label="#eeeee;value:79"

tatcli msg add /Internal/YourTopic/Dashing "#monitoring #YourTeam #item:OCO_Warn" \
  --label="#e8f1f4;bg-color" \
  --label="#fa6800;color" \
  --label="#eeeee;order:12" \
  --label="#eeeee;value:2312"

tatcli msg add /Internal/YourTopic/Dashing "#monitoring #YourTeam #item:OCO_Info" \
  --label="#e8f1f4;bg-color" \
  --label="#60a917;color" \
  --label="#eeeee;order:13" \
  --label="#eeeee;value:2714"

tatcli msg add /Internal/YourTopic/Dashing "#monitoring #YourTeam #item:StatusCake_Tests" \
  --label="#e8f1f4;bg-color" \
  --label="#63305a;color" \
  --label="#eeeee;order:15" \
  --label="#eeeee;value:15/15"

tatcli msg add /Internal/YourTopic/Dashing "#monitoring #YourTeam #item:checkFilerz A" \
  --label="#eeeee;order:18" \
  --label="#e8f1f4;bg-color" \
  --label='#eeeeee;widget:progressbar' \
  --label='#eeeeee;percentRunning:70' \
  --label='#1ba1e2;color' \
  --label='#eeeeee;widget-class:success' \
  --label='#eeeeee;widget-mode:vertical'

tatcli msg add /Internal/YourTopic/Dashing "#monitoring #YourTeam #item:checkFilerz B" \
  --label="#eeeee;order:19" \
  --label="#e8f1f4;bg-color" \
  --label='#eeeeee;widget:progressbar' \
  --label='#eeeeee;percentRunning:30' \
  --label='#1ba1e2;color' \
  --label='#eeeeee;widget-class:success' \
  --label='#eeeeee;widget-mode:horizontal'

tatcli msg add /Internal/YourTopic/Dashing "#monitoring #YourTeam #item:checkFilerz C" \
  --label="#eeeee;order:20" \
  --label="#e8f1f4;bg-color" \
  --label='#eeeeee;widget:progressbar' \
  --label='#eeeeee;percentRunning:25' \
  --label='#eeeeee;widget-class:warning' \
  --label='#eeeeee;widget-mode:horizontal'

tatcli msg add /Internal/YourTopic/Dashing "#monitoring #YourTeam #item:StatusUP" \
  --label="#eeeee;order:21" \
  --label="#60a917;bg-color" \
  --label="#ffffff;color" \
  --label="#eeeeee;value:79"

tatcli msg add /Internal/YourTopic/Dashing "#monitoring #YourTeam #item:Status" \
  --label="#eeeee;order:22" \
  --label="#60a917;bg-color" \
  --label="#ffffff;color" \
  --label="#eeeeee;value::)"

tatcli msg add /Internal/YourTopic/Dashing "#monitoring #YourTeam #item:Status" \
  --label="#eeeee;order:22" \
  --label="#60a917;bg-color" \
  --label="#ffffff;color" \
  --label="#eeeeee;value:↑"

tatcli msg add /Internal/YourTopic/Dashing "#monitoring #item:Pie" \
  --label="#eeeee;order:23" \
  --label="#e8f1f4;bg-color"  \
  --label="#eeeeee;width:20%" \
  --label='#eeeee;widget-data-serie:20 30 40' \
  --label='#eeeee;widget-options:donut:true donutWidth:60 startAngle:270 total:200 showLabel:false' \
  --label='#eeeee;widget:Pie'

tatcli msg add /Internal/YourTopic/Dashing "#monitoring #item:Line" \
  --label="#eeeee;order:24" \
  --label="#e8f1f4;bg-color" \
  --label="#eeeeee;width:76%" \
  --label='#eeeeee;widget-data-labels:Mon Tue Wed Thu Fri Sat' \
  --label='#eeeeee;widget-data-options:low:0 showArea:true showPoint:false fullWidth:true' \
  --label='#eeeeee;widget-data-series:1 5 2 5 4 3' \
  --label='#eeeeee;widget-data-series:2 3 4 8 1 2' \
  --label='#eeeeee;widget-data-series:5 4 3 2 1 0.5' \
  --label='#eeeeee;widget:Line'

tatcli msg add /Internal/YourTopic/Dashing "#monitoring #item:Bar"  \
  --label="#eeeee;order:25" \
  --label="#e8f1f4;bg-color" \
  --label="#eeeeee;height:200px" \
  --label="#eeeeee;width:95%" \
  --label="#eeeeee;widget:Bar" \
  --label="#eeeeee;widget-data-labels:AA BB CC DD EE FF GG" \
  --label="#eeeeee;widget-data-series:5 4 3 7 5 10 3" \
  --label="#eeeeee;widget-data-series:3 2 9 5 4 6 4" \
  --label="#eeeeee;widget-options:seriesBarDistance:10 reverseData:true horizontalBars:true"
```

In standard View, theses messages look like:

![Dashing View in Standard View](/imgs/tatwebui-dashing-view-standardview1.png?width=80%)
![Dashing View in Standard View](/imgs/tatwebui-dashing-view-standardview2.png?width=80%)

### Supported styles

You can add your custom styles by adding labels to your message.
Two types of style syntax are supported: an old syntax which is here for legacy support
and a new syntax you should use from now on which can support almost every CSS style you want.

#### Legacy support (old syntax)

The dashing view provides legacy support for these properties (format: property-name: `syntax` - description):
- bg-color: `bg-color` - When this property is present it will set the background color of the message as the label color
- title-font-size: `title-font-size:[XX]px` - Set the title font size
- value-font-size: `value-font-size:[XX]px` - Set the value font size
- color: `color` - When this property is present it will set the color of the message as the label color
- hide-bottom: `hide-bottom` - When this property is present it will hide the bottom of the message
- url: `url` - When this property is present it will change the cursor to a pointer on the message
- height: `height:[XX]px` - Set message height
- width: `width:[XX]px` - Set message width
- border-width: `border-width:[XX]px` - Set message border-width
- border-style: `border-style:[style]` - Set message border-style
- border-color: `border-color:[color]` - Set message border-color

#### Modern syntax

You can add almost every CSS property you want by following one of these two models:
1. Apply a property to the message: `style:property: value;`
2. Apply a property to a child node of the message: `style:selector { property: value; }`
(`selector` must be set from the message's point of view: if you want to target the div > h3 of your message your selector
must be `div > h3` and TaT Dashing will handle the rest to link this to the right message)

Important notes:
- Do not forget the `;` in those two models or your label will not be processed
- Do not put a space character ` ` after `style:` or your label will not be processed
- Always list one property in a label (for example do not make these labels : `property1: value1; property2: value2;`
or `selector { property1: value1; property2: value2; }`) or your label will not be processed.
Make multiple labels if you want to apply multiple properties.

### Legend

You can add legend on widget-data-series.

```
widget-data-series:LegenA:5 4 3 7 5 10 3
widget-data-series:LegenB:5 4 3 7 5 10 3
```

## Configuration
In plugin.tpl.json file, add this line :

```
"tatwebui-plugin-dashingview": "git+https://github.com/ovh/tatwebui-plugin-dashingview.git"
```

## Source
[https://github.com/ovh/tatwebui-plugin-dashingview](https://github.com/ovh/tatwebui-plugin-dashingview)
