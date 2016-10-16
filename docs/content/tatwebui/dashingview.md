---
weight: 6
title: "Dashing View"
toc: true
prev: "/tatwebui/pastatview"
next: "/tatwebui/releaseview"

---


## Usage

This screenshot

![Dashing View](/imgs/tatwebui-dashing-view.png?width=80%)

can be done by sending info with tatcli :

```bash
tatcli msg add /Inernal/YourTopic/Dashing "#monitoring" --label="#eeeee;border-width:0px,#eeeee;height:110px,#eeeee;hide-bottom,#eeeee;order:1,#eeeee;value:MyDashing,#eeeee;width:96%,#ffffff;color,#161d23;bg-color"
tatcli msg add /Inernal/YourTopic/Dashing "#monitoring" --label="#93c47d;bg-color,#eeeee;border-width:0px,#eeeee;height:5px,#eeeee;hide-bottom,#eeeee;order:2,#eeeee;value,#eeeee;value,#eeeee;value,#eeeee;color,#eeeee;width:96.3%"
tatcli msg add /Inernal/YourTopic/Dashing "#monitoring" --label="#161d23;bg-color,#eeeee;border-width:0px,#eeeee;color,#eeeee;height:20px,#eeeee;hide-bottom,#eeeee;order:3,#eeeee;width:96%"
tatcli msg add /Inernal/YourTopic/Dashing "#monitoring #YourTeam #item:OCO_Critical" --label="#e8f1f4;bg-color,#ce352c;color,#eeeee;order:11,#eeeee;value:79"
tatcli msg add /Inernal/YourTopic/Dashing "#monitoring #YourTeam #item:OCO_Warn" --label="#e8f1f4;bg-color,#fa6800;color,#eeeee;order:12,#eeeee;value:2312"
tatcli msg add /Inernal/YourTopic/Dashing "#monitoring #YourTeam #item:OCO_Info" --label="#e8f1f4;bg-color,#60a917;color,#eeeee;order:13,#eeeee;value:2714"
tatcli msg add /Inernal/YourTopic/Dashing "#monitoring #YourTeam #item:StatusCake_Tests" --label="#e8f1f4;bg-color,#63305a;color,#eeeee;order:15,#eeeee;value:15/15"

tatcli msg add /Inernal/YourTopic/Dashing "#monitoring #YourTeam #item:checkFilerz A" --label="#eeeee;order:18" --label="#e8f1f4;bg-color" --label='#eeeeee;widget:progressbar' --label='#eeeeee;percentRunning:70' --label='#1ba1e2;color' --label='#eeeeee;widget-class:success' --label='#eeeeee;widget-mode:vertical'
tatcli msg add /Inernal/YourTopic/Dashing "#monitoring #YourTeam #item:checkFilerz B" --label="#eeeee;order:19" --label="#e8f1f4;bg-color" --label='#eeeeee;widget:progressbar' --label='#eeeeee;percentRunning:30' --label='#1ba1e2;color' --label='#eeeeee;widget-class:success' --label='#eeeeee;widget-mode:horizontal'
tatcli msg add /Inernal/YourTopic/Dashing "#monitoring #YourTeam #item:checkFilerz C" --label="#eeeee;order:20" --label="#e8f1f4;bg-color" --label='#eeeeee;widget:progressbar' --label='#eeeeee;percentRunning:25' --label='#eeeeee;widget-class:warning' --label='#eeeeee;widget-mode:horizontal'
tatcli msg add /Inernal/YourTopic/Dashing "#monitoring #YourTeam #item:StatusUP" --label="#eeeee;order:21" --label="#60a917;bg-color" --label="#ffffff;color" --label="#eeeeee;value:79"
tatcli msg add /Inernal/YourTopic/Dashing "#monitoring #YourTeam #item:Status" --label="#eeeee;order:22" --label="#60a917;bg-color" --label="#ffffff;color" --label="#eeeeee;value::)"
tatcli msg add /Inernal/YourTopic/Dashing "#monitoring #YourTeam #item:Status" --label="#eeeee;order:22" --label="#60a917;bg-color" --label="#ffffff;color" --label="#eeeeee;value:↑"
tatcli msg add /Inernal/YourTopic/Dashing "#monitoring #item:Pie" --label="#eeeee;order:23" --label="#e8f1f4;bg-color"  --label="#eeeeee;width:20%" --label='#eeeee;widget-data-serie:20 30 40' --label='#eeeee;widget-options:donut:true donutWidth:60 startAngle:270 total:200 showLabel:false' --label='#eeeee;widget:Pie'

tatcli msg add /Inernal/YourTopic/Dashing "#monitoring #item:Line" --label="#eeeee;order:24" --label="#e8f1f4;bg-color" --label="#eeeeee;width:76%" --label='#eeeeee;widget-data-labels:Mon Tue Wed Thu Fri Sat' --label='#eeeeee;widget-data-options:low:0 showArea:true showPoint:false fullWidth:true' --label='#eeeeee;widget-data-series:1 5 2 5 4 3' --label='#eeeeee;widget-data-series:2 3 4 8 1 2' --label='#eeeeee;widget-data-series:5 4 3 2 1 0.5' --label='#eeeeee;widget:Line'

tatcli msg add /Inernal/YourTopic/Dashing "#monitoring #item:Bar"  --label="#eeeee;order:25" --label="#e8f1f4;bg-color" --label="#eeeeee;height:200px" --label="#eeeeee;width:95%" --label="#eeeeee;widget:Bar" --label="#eeeeee;widget-data-labels:AA BB CC DD EE FF GG" --label="#eeeeee;widget-data-series:5 4 3 7 5 10 3" --label="#eeeeee;widget-data-series:3 2 9 5 4 6 4" --label="#eeeeee;widget-options:seriesBarDistance:10 reverseData:true horizontalBars:true"
```

In standard View, theses messages look like:

![Dashing View in Standard View](/imgs/tatwebui-dashing-view-standardview1.png?width=80%)
![Dashing View in Standard View](/imgs/tatwebui-dashing-view-standardview2.png?width=80%)

## Configuration
In plugin.tpl.json file, add this line :

```
"tatwebui-plugin-dashingview": "git+https://github.com/ovh/tatwebui-plugin-dashingview.git"
```
