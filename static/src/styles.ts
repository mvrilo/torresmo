const info = {
  color: "#888"
};

const listItem = {
  minHeight: "40px",
  position: "relative",
  marginBottom: "12px"
};

const listItemBackground = {
  backgroundColor: "darkgreen", 
  position: "absolute",
  minHeight: "50px",
  width: "100%",
  opacity: 0.2,
  left: 0,
  top: 0
};

const listItemContent = {
  background: "transparent",
  position: "absolute",
  minHeight: "40px",
  padding: "5px",
  top: 0
};

const listItemLeft = {
  ...listItemContent,
  width: "75%",
  left: 0
};

const listItemRight = {
  ...listItemContent,
  textAlign: "right",
  width: "%15",
  right: 0
};

export default { info, listItem, listItemBackground, listItemLeft, listItemRight };
