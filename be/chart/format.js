
var createLabelFormatter = () => {
    let callCount = 0;
    return (y) => {
        callCount++;
        if (callCount % 6 === 0) {
        // if (typeof y !== "undefined") {
        //   return y.toFixed(0) + " points";
        // }
        return y;
        }
        return ""; // Return undefined or some other placeholder value
    };
};
