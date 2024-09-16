
var createLabelFormatter = (num) => {
    if (!!!num) {
        num = 6;
    }
    console.log(num)
    let callCount = 0;
    return (y) => {
        callCount++;
        if (callCount % num === 0) {
            // if (typeof y !== "undefined") {
            //   return y.toFixed(0) + " points";
            // }
            return y;
        }
        return ""; // Return undefined or some other placeholder value
    };
};
